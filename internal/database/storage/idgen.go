package storage

import (
	"math"
	"sync/atomic"
)

type IDGen struct {
	counter atomic.Int64
}

func NewIDGen(previousID int64) *IDGen {
	generator := &IDGen{}
	generator.counter.Store(previousID)
	return generator
}

func (g *IDGen) Generate() int64 {
	g.counter.CompareAndSwap(math.MaxInt64, 0)
	return g.counter.Add(1)
}
