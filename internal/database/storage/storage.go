package storage

import (
	"context"
	"fmt"
	"kvdb/internal/conc"
	dwal "kvdb/internal/database/dummy/wal"
)

type Engine interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string)
	Del(ctx context.Context, key string)
}

type WAL interface {
	Set(ctx context.Context, key, value string) conc.FutureError
	Del(ctx context.Context, key string) conc.FutureError
}

type Storage struct {
	engine Engine
	wal    WAL
}

func New(engine Engine) *Storage {
	return &Storage{
		engine: engine,
		wal:    dwal.New(),
	}
}

func (s *Storage) WithWAL(wal WAL) *Storage {
	s.wal = wal
	return s
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	value, _ := s.engine.Get(ctx, key)
	return value, nil
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	ferr := s.wal.Set(ctx, key, value)
	if err := ferr.Get(); err != nil {
		return fmt.Errorf("failed write WAL: %w", err)
	}

	s.engine.Set(ctx, key, value)

	return nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	ferr := s.wal.Del(ctx, key)
	if err := ferr.Get(); err != nil {
		return fmt.Errorf("failed write WAL: %w", err)
	}

	s.engine.Del(ctx, key)
	return nil
}
