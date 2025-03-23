package server

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig_DefaultValues tests loading config with default values.
func TestLoadConfig_DefaultValues(t *testing.T) {
	config := &Config{}
	config.setDefaults()

	assert.Equal(t, "in_memory", config.Engine.Type)
	assert.Equal(t, "127.0.0.1:8080", config.Network.Address)
	assert.Equal(t, 50, config.Network.MaxConnections)
	assert.Equal(t, "2KB", config.Network.MaxMessageSize)
	assert.Equal(t, uint64(2*1024), config.Network.MaxMessageSizeBytes)
	assert.Equal(t, 1*time.Minute, config.Network.IdleTimeout)
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "/var/log/app.log", config.Logging.Output)
}

// TestLoadConfig_FromYAML tests loading config from a YAML file.
func TestLoadConfig_FromYAML(t *testing.T) {
	yamlData := `
engine:
  type: "redis"
network:
  address: "0.0.0.0:8081"
  max_connections: 100
  max_message_size: "4KB"
  idle_timeout: "2m"
logging:
  level: "debug"
  output: "/var/log/debug.log"
`

	reader := bytes.NewBufferString(yamlData)
	config, err := LoadConfig(reader)
	require.NoError(t, err)

	assert.Equal(t, "redis", config.Engine.Type)
	assert.Equal(t, "0.0.0.0:8081", config.Network.Address)
	assert.Equal(t, 100, config.Network.MaxConnections)
	assert.Equal(t, "4KB", config.Network.MaxMessageSize)
	assert.Equal(t, uint64(4000), config.Network.MaxMessageSizeBytes)
	assert.Equal(t, 2*time.Minute, config.Network.IdleTimeout)
	assert.Equal(t, "debug", config.Logging.Level)
	assert.Equal(t, "/var/log/debug.log", config.Logging.Output)
}

// TestLoadConfig_ReadError tests handling of a read error.
func TestLoadConfig_ReadError(t *testing.T) {
	// Mock reader that returns an error
	mockReader := &errorReader{err: errors.New("read error")}

	_, err := LoadConfig(mockReader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed read file")
}

// TestLoadConfig_ParseYAMLError tests handling of a YAML parsing error.
func TestLoadConfig_ParseYAMLError(t *testing.T) {
	invalidYAML := `
engine:
  type: "redis"
network:
  address: "0.0.0.0:8081"
  max_connections: "invalid"  # Invalid type (should be int)
`

	reader := bytes.NewBufferString(invalidYAML)
	_, err := LoadConfig(reader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed parse yaml")
}

// TestLoadConfig_ParseBytesError tests handling of an error when parsing max_message_size.
func TestLoadConfig_ParseBytesError(t *testing.T) {
	invalidYAML := `
network:
  max_message_size: "invalid"
`

	reader := bytes.NewBufferString(invalidYAML)
	_, err := LoadConfig(reader)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed parse bytes")
}

// errorReader is a mock io.Reader that always returns an error.
type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (int, error) {
	return 0, e.err
}
