package main

import (
	"bufio"
	"context"
	"flag"
	cli "kvdb/internal/cli/client"
	"kvdb/internal/network/client"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "database address")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mainLogger := zap.NewExample()

	client, err := initClient(*addr)
	if err != nil {
		mainLogger.Fatal("failed init client", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	reader := bufio.NewReader(os.Stdin)

	// Start CLI.
	cli.Run(ctx, reader, client)
}

func initClient(addr string) (*client.TCPClient, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return client.New(conn), nil
}
