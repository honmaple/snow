package loader

import (
	"iter"
)

type Set[T any] struct {
	list  []T
	index map[string]T
}

func (s *Set[T]) List() []T {
	return s.list
}

func (s *Set[T]) Iter() iter.Seq2[int, T] {
	return func(yield func(key int, value T) bool) {
		for k, v := range s.list {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *Set[T]) Add(key string, val T) {
	if _, ok := s.index[key]; !ok {
		s.list = append(s.list, val)
		s.index[key] = val
	}
}

func (s *Set[T]) Find(key string) (T, bool) {
	value, ok := s.index[key]
	return value, ok
}

func newSet[T any]() *Set[T] {
	return &Set[T]{
		list:  make([]T, 0),
		index: make(map[string]T),
	}
}
