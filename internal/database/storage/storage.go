package storage

import (
	"context"
	"errors"
	"fmt"
	"kvdb/internal/conc"
	"kvdb/internal/database"
	dwal "kvdb/internal/database/dummy/wal"
	"sort"
	"sync"
)

type ctxKey string

const (
	ctxKeyTransactionID ctxKey = "tid"
)

var ErrKeyNotFound = errors.New("key not found")

type Engine interface {
	Get(ctx context.Context, key string) (string, bool)
	Set(ctx context.Context, key, value string)
	Del(ctx context.Context, key string)
}

type Replication interface {
	GetStream() <-chan database.Segment
}

type WAL interface {
	Set(ctx context.Context, key, value string) conc.FutureError
	Del(ctx context.Context, key string) conc.FutureError
	Recover() ([]database.WALRecord, error)
}

type Storage struct {
	engine      Engine
	wal         WAL
	idgen       *IDGen
	replication Replication
	gracefulWg  sync.WaitGroup
}

func New(engine Engine, options ...Option) (*Storage, error) {
	storage := &Storage{
		engine: engine,
		wal:    dwal.New(),
		idgen:  NewIDGen(0),
	}

	for _, option := range options {
		option(storage)
	}

	logs, err := storage.wal.Recover()
	if err != nil {
		return nil, fmt.Errorf("failed to recover: %w", err)
	}

	lastLSN := storage.applyData(logs)
	storage.idgen = NewIDGen(lastLSN)

	if storage.replication != nil {
		storage.gracefulWg.Add(1)
		go func() {
			defer storage.gracefulWg.Done()

			storage.runReplication(storage.replication.GetStream())
		}()
	}

	return storage, nil
}

func (s *Storage) Wait() {
	s.gracefulWg.Wait()
}

func (s *Storage) runReplication(ch <-chan database.Segment) {
	for segment := range ch {
		s.applyData(segment.Logs)
	}
}

func (s *Storage) applyData(logs []database.WALRecord) int64 {
	if len(logs) == 0 {
		return 0
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].LSN < logs[j].LSN
	})

	var lastLSN int64
	for _, log := range logs {
		lastLSN = max(lastLSN, log.LSN)
		ctx := WithTransactionID(context.Background(), s.idgen.Generate())

		switch log.Command {
		case database.CommandSET:
			s.engine.Set(ctx, log.Arguments[0], log.Arguments[1])
		case database.CommandDEL:
			s.engine.Del(ctx, log.Arguments[0])
		}
	}

	return lastLSN
}

func (s *Storage) WithWAL(wal WAL) *Storage {
	s.wal = wal
	return s
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	ctx = WithTransactionID(ctx, s.idgen.Generate())

	value, ok := s.engine.Get(ctx, key)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}

	return value, nil
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	ctx = WithTransactionID(ctx, s.idgen.Generate())

	ferr := s.wal.Set(ctx, key, value)
	if err := ferr.Get(); err != nil {
		return fmt.Errorf("failed write WAL: %w", err)
	}

	s.engine.Set(ctx, key, value)

	return nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	ctx = WithTransactionID(ctx, s.idgen.Generate())

	ferr := s.wal.Del(ctx, key)
	if err := ferr.Get(); err != nil {
		return fmt.Errorf("failed write WAL: %w", err)
	}

	s.engine.Del(ctx, key)
	return nil
}

func WithTransactionID(ctx context.Context, tid int64) context.Context {
	return context.WithValue(ctx, ctxKeyTransactionID, tid)
}

func GetTransactionID(ctx context.Context) (int64, bool) {
	tid, ok := ctx.Value(ctxKeyTransactionID).(int64)
	return tid, ok
}
