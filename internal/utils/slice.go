package utils

import (
	"iter"
)

type Slice[T any] struct {
	list  []T
	index map[string]T
}

func (s *Slice[T]) List() []T {
	return s.list
}

func (s *Slice[T]) Iter() iter.Seq2[int, T] {
	return func(yield func(key int, value T) bool) {
		for k, v := range s.list {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *Slice[T]) Add(key string, val T) {
	s.list = append(s.list, val)
	s.index[key] = val
}

func (s *Slice[T]) Find(key string) (T, bool) {
	u, ok := s.index[key]
	return u, ok
}

func NewSlice[T any]() *Slice[T] {
	return &Slice[T]{
		list:  make([]T, 0),
		index: make(map[string]T),
	}
}
