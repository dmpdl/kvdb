package query

import (
	"context"
	"fmt"
	"strings"
)

type Database interface {
	RunCommand(ctx context.Context, rawQuery string) (string, error)
}

type Handler struct {
	database Database
}

func New(database Database) *Handler {
	return &Handler{
		database: database,
	}
}

func (h *Handler) Handle(ctx context.Context, request []byte) []byte {
	query := strings.TrimSpace(string(request))
	result, err := h.database.RunCommand(ctx, query)

	if err != nil {
		return []byte(fmt.Sprintf("ERR: %s", err.Error()))
	}

	return []byte(result)
}
