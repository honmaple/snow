package core

import (
	"context"
	"io"
	"sync"
)

type (
	Writer interface {
		Write(context.Context, string, io.Reader) error
	}
	Builder interface {
		Build(context.Context) error
	}
)

func Build(ctx *Context, bs ...Builder) error {
	var wg sync.WaitGroup
	for _, b := range bs {
		wg.Add(1)
		go func(builder Builder) {
			defer wg.Done()
			if err := builder.Build(ctx); err != nil {
				ctx.Logger.Errorf("build err: %s", err.Error())
			}
		}(b)
	}
	wg.Wait()
	return nil
}
