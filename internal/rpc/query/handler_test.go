package query

import (
	"bufio"
	"context"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// MockDatabase is a mock implementation of the Database interface for testing.
type MockDatabase struct {
	response string
}

func (m *MockDatabase) RunCommand(_ context.Context, _ string) string {
	return m.response
}

func TestHandler_Handle(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDB := &MockDatabase{response: "mock response\n"}
	handler := New(mockDB, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Start the handler in a goroutine
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.Handle(ctx, serverConn)
	}()

	query := "test query\n"
	_, err := clientConn.Write([]byte(query))
	require.NoError(t, err)

	cancel()

	reader := bufio.NewReader(clientConn)
	response, err := reader.ReadString('\n')
	require.NoError(t, err)

	expectedResponse := "mock response"
	require.Equal(t, expectedResponse, strings.TrimSpace(response))

	wg.Wait()
}

func TestHandler_Handle_ReadError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDB := &MockDatabase{response: "mock response\n"}
	handler := New(mockDB, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Close the client connection to simulate a read error
	clientConn.Close()

	// Start the handler in a goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		handler.Handle(ctx, serverConn)
	}()

	wg.Wait()
}

func TestHandler_Handle_WriteError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockDB := &MockDatabase{response: "mock response\n"}
	handler := New(mockDB, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a pipe to simulate a network connection
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	// Start the handler in a goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		handler.Handle(ctx, serverConn)
	}()

	// Write a query to the client connection
	query := "test query\n"
	_, err := clientConn.Write([]byte(query))
	if err != nil {
		t.Fatalf("Failed to write query to connection: %v", err)
	}

	// Close the client connection to simulate a write error
	clientConn.Close()

	wg.Wait()
}
