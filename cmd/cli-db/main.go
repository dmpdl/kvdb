package main

import (
	"bufio"
	"context"
	cli "kvdb/internal/cli/db"
	"kvdb/internal/compute"
	"kvdb/internal/database"
	"kvdb/internal/storage/inmemory"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Init database.
	logger := zap.NewExample()
	compute := compute.New()
	storage := inmemory.New()
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
