package client

import (
	"context"
	"fmt"
	"net"
)

const (
	defaultBufferSize = 2 * 1024 // 2KB
)

type TCPClient struct {
	conn net.Conn
	opts opts
}

type opts struct {
	bufferSize int // Buffer size, default 2KB.
}

func New(conn net.Conn) *TCPClient {
	return &TCPClient{
		conn: conn,
		opts: opts{
			bufferSize: defaultBufferSize,
		},
	}
}

func (c *TCPClient) WithBufferSize(bufferSize int) *TCPClient {
	c.opts.bufferSize = bufferSize
	return c
}

func (c *TCPClient) Send(_ context.Context, request []byte) ([]byte, error) {
	if len(request) == 0 {
		return []byte{}, nil
	}

	if request[len(request)-1] != '\n' {
		request = append(request, '\n')
	}

	if _, err := c.conn.Write(request); err != nil {
		return []byte{}, fmt.Errorf("failed write conn: %w", err)
	}

	responseBuf := make([]byte, c.opts.bufferSize)
	responseSize, err := c.conn.Read(responseBuf)
	if err != nil {
		return []byte{}, fmt.Errorf("failed read conn: %w", err)
	}

	return responseBuf[:responseSize], nil
}

func (c *TCPClient) Close() error {
	return c.conn.Close()
}
