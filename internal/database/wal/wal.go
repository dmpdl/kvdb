package wal

import (
	"context"
	"kvdb/internal/conc"
	"kvdb/internal/database"
	"sync"
	"time"

	"go.uber.org/zap"
)

type WAL struct {
	logger     *zap.Logger
	opts       opts
	batch      []database.WriteRequest
	mutex      sync.Mutex
	batches    chan []database.WriteRequest
	logsWriter LogsWriter
}

type opts struct {
	flushingBatchLength  int
	flushingBatchTimeout time.Duration
}

type LogsWriter interface {
	Write(batch []database.WriteRequest) error
}

func New(logsWriter LogsWriter, logger *zap.Logger) *WAL {
	return &WAL{
		logsWriter: logsWriter,
		logger:     logger,
	}
}

func (w *WAL) WithFlushingBatchLength(flushingBatchLength int) *WAL {
	w.opts.flushingBatchLength = flushingBatchLength
	return w
}

func (w *WAL) WithFlushingBatchTimeout(flushingBatchTimeout time.Duration) *WAL {
	w.opts.flushingBatchTimeout = flushingBatchTimeout
	return w
}

func (w *WAL) RunFlushing(ctx context.Context) {
	ticker := time.NewTicker(w.opts.flushingBatchTimeout)
	defer ticker.Stop()

	for {
		// Check context closed in first priority.
		select {
		case <-ctx.Done():
			w.logger.Debug("flash batch context closed")
			w.flushBatch()
			return
		default:
		}

		select {
		case <-ticker.C:
			w.logger.Debug("flash batch timeout")

			w.flushBatch()

		case batch := <-w.batches:
			w.logger.Debug("write batch")

			if err := w.logsWriter.Write(batch); err != nil {
				w.logger.Error("failed write batch", zap.Error(err))
			}

			ticker.Reset(w.opts.flushingBatchTimeout)
		}
	}
}

func (w *WAL) Set(ctx context.Context, key, value string) conc.FutureError {
	return w.push(ctx, database.CommandDEL, []string{key, value})
}

func (w *WAL) Del(ctx context.Context, key string) conc.FutureError {
	return w.push(ctx, database.CommandDEL, []string{key})
}

func (w *WAL) push(_ context.Context, command database.Command, args []string) conc.FutureError {
	record := database.NewWriteRequest(command, args)

	conc.WithLock(&w.mutex, func() {
		w.batch = append(w.batch, record)
		if len(w.batch) == w.opts.flushingBatchLength {
			w.batches <- w.batch
			w.batch = nil
		}
	})

	return record.FutureResponse()
}

func (w *WAL) flushBatch() {
	var batch []database.WriteRequest
	conc.WithLock(&w.mutex, func() {
		batch = w.batch
		w.batch = nil
	})

	if len(batch) == 0 {
		return
	}

	for {
		if err := w.logsWriter.Write(batch); err != nil {
			w.logger.Error("failed write batch", zap.Error(err))
			continue
		}

		break
	}
}
