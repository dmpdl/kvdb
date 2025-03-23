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
}

type EngineConfig struct {
	Type string `yaml:"type"`
}

type NetworkConfig struct {
	Address             string        `yaml:"address"`
	MaxConnections      int           `yaml:"max_connections"`
	MaxMessageSize      string        `yaml:"max_message_size"`
	MaxMessageSizeBytes uint64        `yaml:"-"`
	IdleTimeout         time.Duration `yaml:"idle_timeout"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"`
}

func (c *Config) setDefaults() {
	c.Engine.Type = "in_memory"
	c.Network.Address = "127.0.0.1:8080"
	c.Network.MaxConnections = 50
	c.Network.MaxMessageSize = "2KB"
	c.Network.MaxMessageSizeBytes = 2048
	c.Network.IdleTimeout = 1 * time.Minute
	c.Logging.Level = "info"
	c.Logging.Output = "/var/log/app.log"
}

func LoadConfig(r io.Reader) (*Config, error) {
	config := &Config{}
	config.setDefaults()

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
	config.Network.MaxMessageSizeBytes = maxMessageSizeBytes

	return config, nil
}
