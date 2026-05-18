package content

import (
	"sync"
)

type Related[T comparable] struct {
	mu    sync.RWMutex
	once  sync.Once
	cur   T
	list  []T
	index map[T]int
}

func (r *Related[T]) init() {
	r.once.Do(func() {
		r.index = make(map[T]int)
		for i, e := range r.list {
			r.index[e] = i
		}
	})
}

func (r *Related[T]) Prev() T {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result T
	idx, ok := r.index[r.cur]
	if !ok || idx == 0 {
		return result
	}
	return r.list[idx-1]
}

func (r *Related[T]) Next() T {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()

	var result T
	idx, ok := r.index[r.cur]
	if !ok || idx == len(r.list)-1 {
		return result
	}
	return r.list[idx+1]
}

func (r *Related[T]) HasPrev() bool {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()

	idx, ok := r.index[r.cur]
	if !ok || idx == 0 {
		return false
	}
	return true
}

func (r *Related[T]) HasNext() bool {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()

	idx, ok := r.index[r.cur]
	if !ok || idx == len(r.list)-1 {
		return false
	}
	return true
}
