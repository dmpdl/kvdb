package server

import (
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	humanize "github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
)

const (
	// Engine constants.
	DefaultEngineType = "in_memory"

	// Network constants.
	DefaultNetworkAddress      = "127.0.0.1:8080"
	DefaultMaxConnections      = 50
	DefaultMaxMessageSize      = "2KB"
	DefaultMaxMessageSizeBytes = 2048
	DefaultNetworkIdleTimeout  = 1 * time.Minute

	// Logging constants.
	DefaultLogLevel  = "info"
	DefaultLogOutput = "/var/log/app.log"

	// WAL (Write-Ahead Log) constants.
	DefaultFlushingBatchLength  = 100
	DefaultFlushingBatchTimeout = 100 * time.Millisecond
	DefaultMaxSegmentSize       = "1KB"
	DefaultMaxSegmentSizeBytes  = 1024
	DefaultDataDirectory        = "./dir"
)

type Config struct {
	Engine  EngineConfig  `yaml:"engine"`
	Network NetworkConfig `yaml:"network"`
	Logging LoggingConfig `yaml:"logging"`
	WAL     *WALConfig    `yaml:"wal,omitempty"`
}

type EngineConfig struct {
	Type string `yaml:"type"`
}

type NetworkConfig struct {
	Address             string        `yaml:"address"`
	MaxConnections      int           `yaml:"max_connections"`
	MaxMessageSize      string        `yaml:"max_message_size"`
	MaxMessageSizeBytes int           `yaml:"-"`
	IdleTimeout         time.Duration `yaml:"idle_timeout"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

type WALConfig struct {
	FlushingBatchLength  int           `yaml:"flushing_batch_length"`
	FlushingBatchTimeout time.Duration `yaml:"flushing_batch_timeout"`
	MaxSegmentSize       string        `yaml:"max_segment_size"`
	MaxSegmentSizeBytes  int           `yaml:"-"`
	DataDirectory        string        `yaml:"data_directory"`
}

func GetDefaultConfig() Config {
	return Config{
		Engine: EngineConfig{
			Type: DefaultEngineType,
		},
		Network: NetworkConfig{
			Address:             DefaultNetworkAddress,
			MaxConnections:      DefaultMaxConnections,
			MaxMessageSize:      DefaultMaxMessageSize,
			MaxMessageSizeBytes: DefaultMaxMessageSizeBytes,
			IdleTimeout:         DefaultNetworkIdleTimeout,
		},
		Logging: LoggingConfig{
			Level:  DefaultLogLevel,
			Output: DefaultLogOutput,
		},
		WAL: &WALConfig{
			FlushingBatchLength:  DefaultFlushingBatchLength,
			FlushingBatchTimeout: DefaultFlushingBatchTimeout,
			MaxSegmentSize:       DefaultMaxSegmentSize,
			MaxSegmentSizeBytes:  DefaultMaxSegmentSizeBytes,
			DataDirectory:        DefaultDataDirectory,
		},
	}
}

func (c *Config) setDefaults() {
	defaultConf := GetDefaultConfig()
	if c.Engine.Type == "" {
		c.Engine.Type = defaultConf.Engine.Type
	}

	// Network
	if c.Network.Address == "" {
		c.Network.Address = "127.0.0.1:8080"
	}
	if c.Network.MaxConnections == 0 {
		c.Network.MaxConnections = 50
	}
	if c.Network.MaxMessageSize == "" {
		c.Network.MaxMessageSize = defaultConf.Network.MaxMessageSize
		c.Network.MaxMessageSizeBytes = defaultConf.Network.MaxMessageSizeBytes
	}
	if c.Network.IdleTimeout == 0 {
		c.Network.IdleTimeout = 1 * time.Minute
	}

	// Logging
	if c.Logging.Level == "" {
		c.Logging.Level = defaultConf.Logging.Level
	}
	if c.Logging.Output == "" {
		c.Logging.Output = defaultConf.Logging.Output
	}

	// WAL
	if c.WAL != nil {
		if c.WAL.FlushingBatchLength == 0 {
			c.WAL.FlushingBatchLength = defaultConf.WAL.FlushingBatchLength
		}
		if c.WAL.FlushingBatchTimeout == 0 {
			c.WAL.FlushingBatchTimeout = defaultConf.WAL.FlushingBatchTimeout
		}
		if c.WAL.MaxSegmentSize == "" {
			c.WAL.MaxSegmentSize = defaultConf.WAL.MaxSegmentSize
			c.WAL.MaxSegmentSizeBytes = defaultConf.WAL.MaxSegmentSizeBytes
		}
		if c.WAL.DataDirectory == "" {
			c.WAL.DataDirectory = defaultConf.WAL.DataDirectory
		}
	}
}

func LoadConfig(r io.Reader) (*Config, error) {
	config := &Config{}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed read file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed parse yaml: %w", err)
	}

	maxMessageSizeBytes, err := humanize.ParseBytes(config.Network.MaxMessageSize)
	if err != nil {
		return nil, fmt.Errorf("failed parse bytes %s: %w", config.Network.MaxMessageSize, err)
	}
	config.Network.MaxMessageSizeBytes, err = safeUint64ToInt(maxMessageSizeBytes)
	if err != nil {
		return nil, err
	}

	if config.WAL != nil {
		maxSegmentSize, err := humanize.ParseBytes(config.WAL.MaxSegmentSize)
		if err != nil {
			return nil, fmt.Errorf("failed parse bytes %s: %w", config.Network.MaxMessageSize, err)
		}
		config.WAL.MaxSegmentSizeBytes, err = safeUint64ToInt(maxSegmentSize)
		if err != nil {
			return nil, err
		}
	}

	config.setDefaults()

	return config, nil
}

func safeUint64ToInt(u uint64) (int, error) {
	if u > math.MaxInt {
		return 0, errors.New("integer overflow: uint64 value too large for int")
	}
	return int(u), nil
}
