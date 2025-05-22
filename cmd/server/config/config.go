package config

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"kvdb/internal/database"
	"kvdb/internal/database/compute"
	"kvdb/internal/database/engine/inmemory"
	"kvdb/internal/database/fileio"
	"kvdb/internal/database/log/reader"
	"kvdb/internal/database/log/writer"
	"kvdb/internal/database/replication/master"
	"kvdb/internal/database/replication/slave"
	"kvdb/internal/database/storage"
	"kvdb/internal/database/wal"
	"kvdb/internal/network/client"
	"kvdb/internal/network/server"

	serverConfig "kvdb/internal/config/server"
)

type Replication interface {
	Run(ctx context.Context)
}

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

func InitDatabase(storage *storage.Storage, logger *zap.Logger) *database.Database {
	compute := compute.New()
	return database.New(cloneLogger(logger, "database"), compute, storage)
}

func InitStorage(logger *zap.Logger, wal *wal.WAL) *storage.Storage {
	engine := inmemory.New()
	opts := []storage.Option{}
	if wal != nil {
		opts = append(opts, storage.WithWAL(wal))
	}
	storage, err := storage.New(engine, opts...)
	if err != nil {
		logger.Fatal("failed init storage", zap.Error(err))
	}

	return storage
}

func InitWALOptional(conf *serverConfig.Config, logger *zap.Logger) *wal.WAL {
	if conf.WAL == nil {
		return nil
	}

	logsWriter := writer.New(
		fileio.NewSegment(conf.Data, conf.WAL.MaxSegmentSizeBytes))
	logsReader := reader.New(
		fileio.NewSegmentProcessor(conf.Data),
	)

	return wal.New(logsWriter, logsReader, cloneLogger(logger, "wal")).
		WithFlushingBatchLength(conf.WAL.FlushingBatchLength).
		WithFlushingBatchTimeout(conf.WAL.FlushingBatchTimeout)
}

func InitServer(conf *serverConfig.Config, logger *zap.Logger, db *database.Database) (*server.TCPServer, error) {
	listener, err := net.Listen("tcp", conf.Network.Address)
	if err != nil {
		return nil, err
	}

	tcpServer := server.New(cloneLogger(logger, "server"), listener).
		WithMaxConn(conf.Network.MaxConnections).
		WithMaxMessageSize(conf.Network.MaxMessageSizeBytes).
		WithIdleTimeout(conf.Network.IdleTimeout)

	return tcpServer, nil
}

func InitReplicationOptional(logger *zap.Logger, conf *serverConfig.Config) (Replication, error) {
	if conf.Replication == nil {
		return nil, nil
	}

	if conf.WAL == nil && conf.Replication.Type == serverConfig.ReplicationTypeMaster {
		return nil, errors.New("wal is require in master mode")
	}

	if conf.WAL != nil && conf.Replication.Type == serverConfig.ReplicationTypeSlave {
		return nil, errors.New("disable wal in slave mode")
	}

	segmentsReader := fileio.NewSegmentsReader(conf.Data)

	if conf.Replication.Type == serverConfig.ReplicationTypeMaster {
		listener, err := net.Listen("tcp", conf.Replication.MasterAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to listen master address: %w", err)
		}

		server := server.New(cloneLogger(logger, "master-replica"), listener)

		return master.New(logger, server, segmentsReader), nil
	}

	if conf.Replication.Type == serverConfig.ReplicationTypeSlave {
		conn, err := net.Dial("tcp", conf.Replication.MasterAddress)
		if err != nil {
			return nil, err
		}

		client := client.New(conn)

		return slave.New(cloneLogger(logger, "slave-replica"), conf.Replication.SyncInterval, client, segmentsReader), nil
	}

	return nil, errors.New("replication type should be master or slave")
}

func cloneLogger(logger *zap.Logger, name string) *zap.Logger {
	return logger.With(zap.String("me", name))
}
