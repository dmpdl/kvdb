package main

import (
	"context"
	"flag"
	"kvdb/cmd/server/config"
	"os"
	"os/signal"
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

	db := config.InitDatabase(logger)

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

	tcpServer.Listen(ctx)
}
