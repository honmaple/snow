package static

import (
	"sync"

	"github.com/honmaple/snow/config"
)

type Context struct {
	mu sync.RWMutex

	files    Statics
	filesMap map[string]*Static
}

func (ctx *Context) Statics() Statics {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	return ctx.files
}

func (ctx *Context) insertStatic(file *Static) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.files = append(ctx.files, file)
	ctx.filesMap[file.Name] = file
}

func newContext(conf config.Config) *Context {
	return &Context{
		filesMap: make(map[string]*Static),
	}
}
