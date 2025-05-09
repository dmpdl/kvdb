package main

import (
	"bufio"
	"context"
	cli "kvdb/internal/cli/db"
	"kvdb/internal/database"
	"kvdb/internal/database/compute"
	"kvdb/internal/database/engine/inmemory"
	"kvdb/internal/database/storage"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Init database.
	logger := zap.NewExample()
	compute := compute.New()
	engine := inmemory.New()
	storage, err := storage.New(engine)
	if err != nil {
		logger.Fatal("failed init storage", zap.Error(err))
	}
	database := database.New(logger, compute, storage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	reader := bufio.NewReader(os.Stdin)

	// Start CLI.
	cli.Run(ctx, reader, database)
}
