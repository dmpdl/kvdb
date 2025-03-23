package server

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

// MockListener is a mock implementation of net.Listener for testing.
type MockListener struct {
	AcceptConn net.Conn
	AcceptErr  error
	CloseErr   error
	Closed     bool
}

func (m *MockListener) Accept() (net.Conn, error) {
	return m.AcceptConn, m.AcceptErr
}

func (m *MockListener) Close() error {
	m.Closed = true
	return m.CloseErr
}

func (m *MockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

// MockConn is a mock implementation of net.Conn for testing.
type MockConn struct {
	ReadBuffer  *bytes.Buffer
	WriteBuffer *bytes.Buffer
	Closed      bool
}

func (m *MockConn) Read(b []byte) (int, error) {
	return m.ReadBuffer.Read(b)
}

func (m *MockConn) Write(b []byte) (int, error) {
	return m.WriteBuffer.Write(b)
}

func (m *MockConn) Close() error {
	m.Closed = true
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *MockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 12345}
}

func (m *MockConn) SetDeadline(_ time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(_ time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(_ time.Time) error {
	return nil
}

// TestNewTCPServer tests the creation of a new TCPServer with default options.
func TestNewTCPServer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := &MockListener{}

	server := New(logger, listener)

	if server.opts.maxConn != defaultMaxConn {
		t.Errorf("Expected maxConn %d, got %d", defaultMaxConn, server.opts.maxConn)
	}
	if server.opts.maxMessageSizeBytes != defaultMaxMessageSizeBytes {
		t.Errorf("Expected maxMessageSizeBytes %d, got %d", defaultMaxMessageSizeBytes, server.opts.maxMessageSizeBytes)
	}
	if server.opts.idleTimeout != defaultIdleTimeout {
		t.Errorf("Expected idleTimeout %v, got %v", defaultIdleTimeout, server.opts.idleTimeout)
	}
}

// TestWithOptions tests setting custom options on the TCPServer.
func TestWithOptions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := &MockListener{}

	server := New(logger, listener).
		WithMaxConn(200).
		WithMaxMessageSize(4096).
		WithIdleTimeout(2 * time.Minute)

	if server.opts.maxConn != 200 {
		t.Errorf("Expected maxConn 200, got %d", server.opts.maxConn)
	}
	if server.opts.maxMessageSizeBytes != 4096 {
		t.Errorf("Expected maxMessageSizeBytes 4096, got %d", server.opts.maxMessageSizeBytes)
	}
	if server.opts.idleTimeout != 2*time.Minute {
		t.Errorf("Expected idleTimeout 2m, got %v", server.opts.idleTimeout)
	}
}

// TestListenLoop tests the listen loop with a mock connection.
func TestListenLoop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString("test query\n"),
		WriteBuffer: new(bytes.Buffer),
	}
	listener := &MockListener{
		AcceptConn: mockConn,
	}

	server := New(logger, listener)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Listen(ctx)
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop the server
	cancel()
	wg.Wait()

	// Verify the connection was closed
	if !mockConn.Closed {
		t.Error("Expected connection to be closed")
	}
}

// TestListenLoop_AcceptError tests handling of an error in Accept.
func TestListenLoop_AcceptError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	listener := &MockListener{
		AcceptErr: errors.New("accept error"),
	}

	server := New(logger, listener)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Listen(ctx)
	}()

	// Wait for the server to start
	time.Sleep(5 * time.Millisecond)

	// Cancel the context to stop the server
	cancel()
	wg.Wait()
}

// TestListenLoop_ContextCanceled tests stopping the server by canceling the context.
func TestListenLoop_ContextCanceled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString("test query\n"),
		WriteBuffer: new(bytes.Buffer),
	}
	listener := &MockListener{AcceptConn: mockConn}

	server := New(logger, listener)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Listen(ctx)
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Cancel the context to stop the server
	cancel()
	wg.Wait()

	// Verify the listener was closed
	if !listener.Closed {
		t.Error("Expected listener to be closed")
	}
}

// TestListenLoop_WithPanic tests panic handler.
func TestListenLoop_WithPanic(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString("test query\n"),
		WriteBuffer: new(bytes.Buffer),
	}
	listener := &MockListener{AcceptConn: mockConn}

	server := New(logger, listener).WithQueryHandleFunc(
		func(_ context.Context, _ net.Conn) {
			panic("test panic")
		})

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Listen(ctx)
	}()

	time.Sleep(5 * time.Millisecond)

	cancel()
	wg.Wait()

	if !listener.Closed {
		t.Error("Expected listener to be closed")
	}
}
