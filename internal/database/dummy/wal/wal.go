package dummywal

import (
	"context"
	"kvdb/internal/conc"
)

// DummyWAL always return nil error.
type DummyWAL struct{}

func New() *DummyWAL {
	return &DummyWAL{}
}

func (w *DummyWAL) Set(ctx context.Context, key, value string) conc.FutureError {
	promise := conc.NewPromise[error]()
	promise.Set(nil)
	return promise.GetFuture()
}

func (w *DummyWAL) Del(ctx context.Context, key string) conc.FutureError {
	promise := conc.NewPromise[error]()
	promise.Set(nil)
	return promise.GetFuture()
}
