package inmemory

import (
	"context"
	"sync"
)

type Engine struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *Engine {
	return &Engine{
		mu:   sync.RWMutex{},
		data: make(map[string]string),
	}
}

func (s *Engine) Get(_ context.Context, key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.data[key]
	return value, ok
}

func (s *Engine) Set(_ context.Context, key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

func (s *Engine) Del(_ context.Context, key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}
