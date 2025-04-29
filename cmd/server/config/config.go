package config

import (
	"fmt"
	"net"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"kvdb/internal/database"
	"kvdb/internal/database/compute"
	"kvdb/internal/database/engine/inmemory"
	"kvdb/internal/database/storage"
	"kvdb/internal/network/server"
	"kvdb/internal/rpc/query"

	serverConfig "kvdb/internal/config/server"
)

func LoadConfig(configPath string) (*serverConfig.Config, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed read file: %s: %w", configPath, err)
	}
	defer f.Close()

	return serverConfig.LoadConfig(f)
}

func InitLogger(config *serverConfig.Config) (*zap.Logger, error) {
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

func InitDatabase(logger *zap.Logger) *database.Database {
	compute := compute.New()
	engine := inmemory.New()
	storage := storage.New(engine)
	return database.New(logger, compute, storage)
}

func InitServer(conf *serverConfig.Config, logger *zap.Logger, db *database.Database) (*server.TCPServer, error) {
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
