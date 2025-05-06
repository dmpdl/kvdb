package server

import (
	"fmt"
	"io"
	"time"

	humanize "github.com/dustin/go-humanize"
	"gopkg.in/yaml.v3"
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
	MaxSegmentSizeBytes  uint64
	DataDirectory        string `yaml:"data_directory"`
}

func GetDefaultConfig() Config {
	return Config{
		Engine: EngineConfig{
			Type: "in_memory",
		},
		Network: NetworkConfig{
			Address:             "127.0.0.1:8080",
			MaxConnections:      50,
			MaxMessageSize:      "2KB",
			MaxMessageSizeBytes: 2048,
			IdleTimeout:         1 * time.Minute,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Output: "/var/log/app.log",
		},
		WAL: &WALConfig{
			FlushingBatchLength:  100,
			FlushingBatchTimeout: 100 * time.Millisecond,
			MaxSegmentSize:       "1KB",
			MaxSegmentSizeBytes:  1024,
			DataDirectory:        "./dir",
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
	config.Network.MaxMessageSizeBytes = int(maxMessageSizeBytes)

	if config.WAL != nil {
		maxSegmentSize, err := humanize.ParseBytes(config.WAL.MaxSegmentSize)
		if err != nil {
			return nil, fmt.Errorf("failed parse bytes %s: %w", config.Network.MaxMessageSize, err)
		}
		config.WAL.MaxSegmentSizeBytes = maxSegmentSize
	}

	config.setDefaults()

	return config, nil
}
