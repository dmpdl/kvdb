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

func (w *DummyWAL) Set(_ context.Context, _, _ string) conc.FutureError {
	promise := conc.NewPromise[error]()
	promise.Set(nil)
	return promise.GetFuture()
}

func (w *DummyWAL) Del(_ context.Context, _ string) conc.FutureError {
	promise := conc.NewPromise[error]()
	promise.Set(nil)
	return promise.GetFuture()
}
