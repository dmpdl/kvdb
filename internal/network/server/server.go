package server

import (
	"context"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultMaxConn             = 100
	defaultMaxMessageSizeBytes = 2 * 1024 // 2KB.
	defaultIdleTimeout         = time.Minute
)

type TCPServer struct {
	logger   *zap.Logger
	listener net.Listener
	handler  handleFunc
	opts     opts
}

type opts struct {
	maxConn             int           // Max number of connections. Default 100.
	maxMessageSizeBytes uint64        // Max message size in bytes. Default 2Kb.
	idleTimeout         time.Duration // Idle timeout. Default 1m.
}

type handleFunc func(ctx context.Context, conn net.Conn)

var dummyHandleFunc handleFunc = func(_ context.Context, conn net.Conn) {
	defer conn.Close()
}

func New(logger *zap.Logger, listener net.Listener) *TCPServer {
	return &TCPServer{
		logger:   logger,
		listener: listener,
		opts: opts{
			maxConn:             defaultMaxConn,
			maxMessageSizeBytes: defaultMaxMessageSizeBytes,
			idleTimeout:         defaultIdleTimeout,
		},
		handler: dummyHandleFunc,
	}
}

func (n *TCPServer) WithMaxConn(maxConn int) *TCPServer {
	n.opts.maxConn = maxConn
	return n
}

func (n *TCPServer) WithMaxMessageSize(maxMessageSizeBytes uint64) *TCPServer {
	n.opts.maxMessageSizeBytes = maxMessageSizeBytes
	return n
}

func (n *TCPServer) WithIdleTimeout(idleTimeout time.Duration) *TCPServer {
	n.opts.idleTimeout = idleTimeout
	return n
}

func (n *TCPServer) WithQueryHandleFunc(handler handleFunc) *TCPServer {
	n.handler = handler
	return n
}

func (n *TCPServer) Listen(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		n.listenLoop(ctx)
	}()

	<-ctx.Done()

	if err := n.listener.Close(); err != nil {
		n.logger.Error(
			"failed to close listener",
			zap.String("addr", n.listener.Addr().String()),
		)
	}

	wg.Wait()
}

func (n *TCPServer) listenLoop(ctx context.Context) {
	n.logger.Info(
		"start serve",
		zap.String("addr", n.listener.Addr().String()),
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := n.listener.Accept()
		if err != nil {
			n.logger.Error(
				"failed to accept conn",
				zap.String("addr", n.listener.Addr().String()),
			)
			continue
		}

		if err := conn.SetReadDeadline(
			time.Now().Add(n.opts.idleTimeout)); err != nil {
			n.logger.Error(
				"failed set idle timeout",
				zap.String("addr", n.listener.Addr().String()),
			)
			continue
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					n.logger.Error(
						"panic",
						zap.String("conn", conn.LocalAddr().String()),
						zap.Any("panic", err),
					)
				}
			}()

			n.handler(ctx, conn)
		}()
	}
}
