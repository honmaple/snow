package site

import (
	"iter"
	"sync"
)

type Set[T any] struct {
	mu    sync.RWMutex
	list  []T
	index map[string]T
}

func (s *Set[T]) List() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.list
}

func (s *Set[T]) Iter() iter.Seq2[int, T] {
	return func(yield func(key int, value T) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		for k, v := range s.list {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *Set[T]) Add(key string, val T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.index[key]; !ok {
		s.list = append(s.list, val)
		s.index[key] = val
	}
}

func (s *Set[T]) Find(key string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, ok := s.index[key]
	return value, ok
}

func newSet[T any]() *Set[T] {
	return &Set[T]{
		list:  make([]T, 0),
		index: make(map[string]T),
	}
}
