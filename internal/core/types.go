package core

import (
	"context"
	"io"
)

type (
	Writer interface {
		Write(context.Context, string, io.Reader) error
	}
	Builder interface {
		Build(context.Context) error
	}
)
