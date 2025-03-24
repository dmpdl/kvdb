package query

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"

	"go.uber.org/zap"
)

type Database interface {
	RunCommand(ctx context.Context, rawQuery string) string
}

type Handler struct {
	database Database
	logger   *zap.Logger
}

func New(database Database, logger *zap.Logger) *Handler {
	return &Handler{
		database: database,
		logger:   logger,
	}
}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		query, err := reader.ReadString('\n')
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				h.logger.Warn("read timeout", zap.Error(err))
				return
			}

			h.logger.Error("failed read conn", zap.Error(err))
			return
		}

		query = strings.TrimSpace(query)
		result := h.database.RunCommand(ctx, query)

		_, err = conn.Write([]byte(result))
		if err != nil {
			h.logger.Error("failed write conn", zap.Error(err))
			return
		}
	}
}
