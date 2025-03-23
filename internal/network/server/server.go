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

func (s *TCPServer) WithMaxConn(maxConn int) *TCPServer {
	s.opts.maxConn = maxConn
	return s
}

func (s *TCPServer) WithMaxMessageSize(maxMessageSizeBytes uint64) *TCPServer {
	s.opts.maxMessageSizeBytes = maxMessageSizeBytes
	return s
}

func (s *TCPServer) WithIdleTimeout(idleTimeout time.Duration) *TCPServer {
	s.opts.idleTimeout = idleTimeout
	return s
}

func (s *TCPServer) WithQueryHandleFunc(handler handleFunc) *TCPServer {
	s.handler = handler
	return s
}

func (s *TCPServer) Listen(ctx context.Context) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		s.listenLoop(ctx)
	}()

	<-ctx.Done()

	if err := s.listener.Close(); err != nil {
		s.logger.Error(
			"failed to close listener",
			zap.String("addr", s.listener.Addr().String()),
		)
	}

	wg.Wait()
}

func (s *TCPServer) listenLoop(ctx context.Context) {
	s.logger.Info(
		"start serve",
		zap.String("addr", s.listener.Addr().String()),
	)

	maxConnCh := make(chan struct{}, s.opts.maxConn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := s.listener.Accept()
		if err != nil {
			s.logger.Error(
				"failed to accept conn",
				zap.String("addr", s.listener.Addr().String()),
			)
			continue
		}

		if err := conn.SetReadDeadline(
			time.Now().Add(s.opts.idleTimeout)); err != nil {
			s.logger.Error(
				"failed set idle timeout",
				zap.String("addr", s.listener.Addr().String()),
			)
			continue
		}

		maxConnCh <- struct{}{}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					s.logger.Error(
						"panic",
						zap.String("conn", conn.LocalAddr().String()),
						zap.Any("panic", err),
					)
				}
			}()

			defer func() { <-maxConnCh }()

			s.handler(ctx, conn)
		}()
	}
}
