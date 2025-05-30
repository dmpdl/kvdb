package main

import (
	"context"
	"flag"
	"kvdb/cmd/server/config"
	"kvdb/internal/rpc/query"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "etc/master.yaml", "config path")
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
		storage = config.InitStorage(logger, wal)
		db      = config.InitDatabase(storage, logger)
	)

	replication, err := config.InitReplicationOptional(logger, conf)
	if err != nil {
		mainLogger.Fatal("failed init replication", zap.Error(err))
	}

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
		tcpServer.ListenFunc(ctx, query.New(db).Handle)
	}()

	// Start WAL flushing
	wg.Add(1)
	go func() {
		defer wg.Done()

		if wal == nil {
			return
		}

		wal.RunFlushing(ctx)
	}()

	// Start Replication process
	wg.Add(1)
	go func() {
		defer wg.Done()

		if replication == nil {
			return
		}

		replication.Run(ctx)
		replication.Wait()
	}()

	wg.Wait()
}
