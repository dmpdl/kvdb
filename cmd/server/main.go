package main

import (
	"context"
	"flag"
	"kvdb/cmd/server/config"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "etc/server.yaml", "config path")
	flag.Parse()

	mainLogger := zap.NewExample()

	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		mainLogger.Fatal("failed load config", zap.Error(err))
	}

	logger, err := config.InitLogger(conf)
	if err != nil {
		mainLogger.Fatal("failed init logger", zap.Error(err))
	}

	var (
		wal     = config.InitWALOptional(conf, logger)
		storage = config.InitStorage(wal)
		db      = config.InitDatabase(storage, logger)
	)

	tcpServer, err := config.InitServer(conf, logger, db)
	if err != nil {
		mainLogger.Fatal("failed init server", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	var wg sync.WaitGroup

	// Start listening
	wg.Add(1)
	go func() {
		defer wg.Done()
		tcpServer.Listen(ctx)
	}()

	// Start WAL flushing
	wg.Add(1)
	go func() {
		defer wg.Done()
		wal.RunFlushing(ctx)
	}()

	wg.Wait()
}
