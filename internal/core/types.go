package core

import (
	"context"
	"fmt"
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

type AsyncBuilder struct {
	Builders []Builder
}

func (ab *AsyncBuilder) Build(ctx context.Context) error {
	var wg sync.WaitGroup
	for _, b := range ab.Builders {
		wg.Add(1)
		go func(builder Builder) {
			defer wg.Done()
			if err := builder.Build(ctx); err != nil {
				fmt.Println(err.Error())
			}
		}(b)
	}
	wg.Wait()
	return nil
}

func Build(ctx context.Context, bs ...Builder) error {
	var wg sync.WaitGroup
	for _, b := range bs {
		wg.Add(1)
		go func(builder Builder) {
			defer wg.Done()
			if err := builder.Build(ctx); err != nil {
				fmt.Println(err.Error())
			}
		}(b)
	}
	wg.Wait()
	return nil
}
