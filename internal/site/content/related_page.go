package content

import (
	"sync"
)

type RelatedPages struct {
	mu    sync.RWMutex
	once  sync.Once
	list  Pages
	index map[*Page]int
}

func (r *RelatedPages) init() {
	r.once.Do(func() {
		r.index = make(map[*Page]int)
		for i, page := range r.list {
			r.index[page] = i
		}
	})
}

func (r *RelatedPages) Prev(page *Page) *Page {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()
	idx, ok := r.index[page]
	if !ok || idx == 0 {
		return nil
	}
	return r.list[idx-1]
}

func (r *RelatedPages) Next(page *Page) *Page {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()
	idx, ok := r.index[page]
	if !ok || idx == len(r.list)-1 {
		return nil
	}
	return r.list[idx+1]
}
