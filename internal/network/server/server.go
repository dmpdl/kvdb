package server

import (
	"context"
	"io"
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
	logger     *zap.Logger
	listener   net.Listener
	handleFunc handleFunc
	opts       opts
}

type opts struct {
	maxConn             int           // Max number of connections. Default 100.
	maxMessageSizeBytes int           // Max message size in bytes. Default 2Kb.
	idleTimeout         time.Duration // Idle timeout. Default 1m.
}

type handleFunc func(ctx context.Context, request []byte) []byte

func New(logger *zap.Logger, listener net.Listener, handleFunc handleFunc) *TCPServer {
	return &TCPServer{
		logger:   logger,
		listener: listener,
		opts: opts{
			maxConn:             defaultMaxConn,
			maxMessageSizeBytes: defaultMaxMessageSizeBytes,
			idleTimeout:         defaultIdleTimeout,
		},
		handleFunc: handleFunc,
	}
}

func (s *TCPServer) WithMaxConn(maxConn int) *TCPServer {
	s.opts.maxConn = maxConn
	return s
}

func (s *TCPServer) WithMaxMessageSize(maxMessageSizeBytes int) *TCPServer {
	s.opts.maxMessageSizeBytes = maxMessageSizeBytes
	return s
}

func (s *TCPServer) WithIdleTimeout(idleTimeout time.Duration) *TCPServer {
	s.opts.idleTimeout = idleTimeout
	return s
}

func (s *TCPServer) Listen(ctx context.Context) {
	wg := sync.WaitGroup{}

	s.logger.Info(
		"start listen loop",
		zap.String("addr", s.listener.Addr().String()),
		zap.Int("max_conn", s.opts.maxConn),
		zap.Int("max_message_size_bytes", s.opts.maxMessageSizeBytes),
		zap.String("idle_timeout", s.opts.idleTimeout.String()),
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.listenLoop(ctx)
	}()

	<-ctx.Done()

	s.logger.Info("close listener")

	if err := s.listener.Close(); err != nil {
		s.logger.Error(
			"failed to close listener",
			zap.String("addr", s.listener.Addr().String()),
		)
	}

	wg.Wait()
}

func (s *TCPServer) listenLoop(ctx context.Context) {
	// Limit max concurrent connections.
	connLimiter := newConnectionLimiter(s.opts.maxConn)

	wg := sync.WaitGroup{}

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		default:
		}

		connLimiter.Acquire()

		conn, err := s.listener.Accept()
		if err != nil {
			connLimiter.Release()
			s.logger.Error(
				"failed to accept conn",
				zap.String("addr", s.listener.Addr().String()),
			)
			continue
		}

		if err := conn.SetReadDeadline(
			time.Now().Add(s.opts.idleTimeout),
		); err != nil {
			connLimiter.Release()
			s.logger.Error(
				"failed set idle timeout",
				zap.String("addr", s.listener.Addr().String()),
			)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			s.handleConn(ctx, conn, connLimiter)
		}()
	}
}

func (s *TCPServer) handleConn(ctx context.Context, conn net.Conn, connLimiter *connectionLimiter) {
	defer func() {
		if err := recover(); err != nil {
			s.logger.Error(
				"panic",
				zap.String("conn", conn.LocalAddr().String()),
				zap.Any("panic", err),
			)
		}

		if err := conn.Close(); err != nil {
			s.logger.Warn("failed to close connection", zap.Error(err))
		}
	}()

	logger := s.logger.With(zap.String("addr", conn.RemoteAddr().String()))

	defer func() {
		connLimiter.Release()
	}()

	// reuse buffer for requests
	request := make([]byte, s.opts.maxMessageSizeBytes)

	for {
		if s.opts.idleTimeout != 0 {
			if err := conn.SetReadDeadline(time.Now().Add(s.opts.idleTimeout)); err != nil {
				s.logger.Warn("failed to set read deadline", zap.Error(err))
				break
			}
		}

		count, err := conn.Read(request)
		if err != nil && err != io.EOF {
			logger.Warn(
				"failed to read data",
				zap.Error(err),
			)
			break
		} else if count == int(s.opts.maxMessageSizeBytes) {
			logger.Warn(
				"small buffer size",
				zap.Int("buffer_size", s.opts.maxMessageSizeBytes),
			)
			break
		}

		if s.opts.idleTimeout != 0 {
			if err := conn.SetWriteDeadline(time.Now().Add(s.opts.idleTimeout)); err != nil {
				s.logger.Warn("failed to set read deadline", zap.Error(err))
				break
			}
		}

		response := s.handleFunc(ctx, request[:count])
		if len(response) == 0 {
			s.logger.Warn("empty response")
		}
		if _, err := conn.Write(response); err != nil {
			s.logger.Warn(
				"failed to write data",
				zap.Error(err),
			)
			break
		}
	}
}

type connectionLimiter struct {
	maxConn   int
	maxConnCh chan struct{}
}

func newConnectionLimiter(maxConn int) *connectionLimiter {
	var maxConnCh chan struct{}
	if maxConn > 0 {
		maxConnCh = make(chan struct{}, maxConn)
	}

	return &connectionLimiter{
		maxConn:   maxConn,
		maxConnCh: maxConnCh,
	}
}

func (cl *connectionLimiter) Release() {
	if cl.maxConnCh != nil {
		<-cl.maxConnCh
	}
}

func (cl *connectionLimiter) Acquire() {
	if cl.maxConnCh != nil {
		cl.maxConnCh <- struct{}{}
	}
}
