package client

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MockConn is a mock implementation of net.Conn for testing.
type MockConn struct {
	ReadBuffer  *bytes.Buffer
	WriteBuffer *bytes.Buffer
	Closed      bool
	ReadError   error
	WriteError  error
}

func (m *MockConn) Read(b []byte) (int, error) {
	if m.ReadError != nil {
		return 0, m.ReadError
	}
	return m.ReadBuffer.Read(b)
}

func (m *MockConn) Write(b []byte) (int, error) {
	if m.WriteError != nil {
		return 0, m.WriteError
	}
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

// TestSend_Success tests successful sending and receiving of data.
func TestSend_Success(t *testing.T) {
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString("response\n"),
		WriteBuffer: new(bytes.Buffer),
	}

	client := New(mockConn)

	request := []byte("request")
	response, err := client.Send(context.Background(), request)
	require.NoError(t, err)

	expectedResponse := []byte("response\n")
	require.Equal(t, expectedResponse, response)

	expectedRequest := []byte("request\n")
	require.Equal(t, expectedRequest, mockConn.WriteBuffer.Bytes())
}

// TestSend_EmptyRequest tests sending an empty request.
func TestSend_EmptyRequest(t *testing.T) {
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString("response\n"),
		WriteBuffer: new(bytes.Buffer),
	}

	client := New(mockConn)

	request := []byte{}
	response, err := client.Send(context.Background(), request)
	require.NoError(t, err)

	require.Empty(t, response)
	require.Equal(t, 0, mockConn.WriteBuffer.Len())
}

// TestSend_WriteError tests handling of a write error.
func TestSend_WriteError(t *testing.T) {
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString("response\n"),
		WriteBuffer: new(bytes.Buffer),
	}

	mockConn.WriteError = errors.New("test error")

	client := New(mockConn)

	request := []byte("request")
	_, err := client.Send(context.Background(), request)
	require.Error(t, err)
	require.ErrorContains(t, err, "test error")
}

// TestSend_ReadError tests handling of a read error.
func TestSend_ReadError(t *testing.T) {
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString(""),
		WriteBuffer: new(bytes.Buffer),
	}

	// Override the Read method to return an error
	mockConn.ReadError = errors.New("test error")
	client := New(mockConn)

	request := []byte("request")
	_, err := client.Send(context.Background(), request)
	require.Error(t, err)
	require.ErrorContains(t, err, "test error")
}

// TestClose tests closing the connection.
func TestClose(t *testing.T) {
	mockConn := &MockConn{
		ReadBuffer:  bytes.NewBufferString(""),
		WriteBuffer: new(bytes.Buffer),
	}

	client := New(mockConn)

	err := client.Close()
	require.NoError(t, err)

	require.True(t, mockConn.Closed)
}
