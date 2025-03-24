package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"kvdb/internal/compute"
	serverConfig "kvdb/internal/config/server"
	"kvdb/internal/database"
	"kvdb/internal/network/server"
	"kvdb/internal/rpc/query"
	"kvdb/internal/storage/inmemory"
)

func main() {
	configPath := flag.String("config", "etc/server.yaml", "config path")
	flag.Parse()

	mainLogger := zap.NewExample()

	config, err := loadConfig(*configPath)
	if err != nil {
		mainLogger.Fatal("failed load config", zap.Error(err))
	}

	logger, err := initLogger(config)
	if err != nil {
		mainLogger.Fatal("failed init logger", zap.Error(err))
	}

	db := initDatabase(logger)

	tcpServer, err := initServer(config, logger, db)
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

func loadConfig(configPath string) (*serverConfig.Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed read file: %s: %w", configPath, err)
	}
	defer f.Close()

	return serverConfig.LoadConfig(f)
}

func initLogger(config *serverConfig.Config) (*zap.Logger, error) {
	var level zapcore.Level
	err := level.UnmarshalText([]byte(config.Logging.Level))
	if err != nil {
		return nil, fmt.Errorf("unexpected log level: %w", err)
	}

	var output zapcore.WriteSyncer
	if config.Logging.Output == "" || config.Logging.Output == "stdout" {
		output = os.Stdout
	} else {
		file, err := os.OpenFile(config.Logging.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed open log file: %w", err)
		}
		output = zapcore.AddSync(file)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		output,
		level,
	)

	logger := zap.New(core)
	return logger, nil
}

func initDatabase(logger *zap.Logger) *database.Database {
	compute := compute.New()
	storage := inmemory.New()
	return database.New(logger, compute, storage)
}

func initServer(conf *serverConfig.Config, logger *zap.Logger, db *database.Database) (*server.TCPServer, error) {
	listener, err := net.Listen("tcp", conf.Network.Address)
	if err != nil {
		return nil, err
	}

	queryHandler := query.New(db, logger)

	tcpServer := server.New(logger, listener).
		WithMaxConn(conf.Network.MaxConnections).
		WithMaxMessageSize(conf.Network.MaxMessageSizeBytes).
		WithIdleTimeout(conf.Network.IdleTimeout).
		WithQueryHandleFunc(queryHandler.Handle)

	return tcpServer, nil
}
