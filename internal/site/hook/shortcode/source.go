package shortcode

import "sync"

type Source interface {
	Id() string
	Content() string
	Context() map[string]any
	Get(string) (any, bool)
	Set(string, any)
}

type source struct {
	id      string
	content string
	context map[string]any
	state   sync.Map
}

func (s *source) Id() string {
	return s.id
}

func (s *source) Content() string {
	return s.content
}

func (s *source) Context() map[string]any {
	return s.context
}

func (s *source) Get(key string) (any, bool) {
	return s.state.Load(key)
}

func (s *source) Set(key string, value any) {
	s.state.Store(key, value)
}
