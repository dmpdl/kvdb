package wal

import (
	"context"
	"kvdb/internal/conc"
	"kvdb/internal/database"
	"sync"
	"time"
)

type WAL struct {
	opts       opts
	batch      []database.WriteRequest
	mutex      sync.Mutex
	batches    chan []database.WriteRequest
	logsWriter LogsWriter
}

type opts struct {
	maxBatchSize int
	batchTimeout time.Duration
}

type LogsWriter interface {
	Write(batch []database.WriteRequest)
}

func New() *WAL {
	return &WAL{}
}

func (w *WAL) WithMaxBatchSize(maxBatchSize int) *WAL {
	w.opts.maxBatchSize = maxBatchSize
	return w
}

func (w *WAL) WithBatchTimeout(batchTimeout time.Duration) *WAL {
	w.opts.batchTimeout = batchTimeout
	return w
}

func (w *WAL) Run(ctx context.Context) {
	ticker := time.NewTicker(w.opts.batchTimeout)
	defer ticker.Stop()

	for {
		// Check context closed in first priority.
		select {
		case <-ctx.Done():
			w.flushBatch()
			return
		default:
		}

		select {
		case <-ticker.C:
			w.flushBatch()

		case batch := <-w.batches:
			w.logsWriter.Write(batch)
			ticker.Reset(w.opts.batchTimeout)
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
		if len(w.batch) == w.opts.maxBatchSize {
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

	w.logsWriter.Write(batch)
}
