package core

import (
	"context"
	"io"
)

type (
	Logger interface {
		Debug(...any)
		Debugf(string, ...any)
		Debugln(...any)
		Info(...any)
		Infof(string, ...any)
		Infoln(...any)
		Warn(...any)
		Warnf(string, ...any)
		Warnln(...any)
		Error(...any)
		Errorf(string, ...any)
		Errorln(...any)
		Fatal(...any)
		Fatalf(string, ...any)
		Fatalln(...any)
	}
	Writer interface {
		Write(context.Context, string, io.Reader) error
	}
	Builder interface {
		Build(context.Context) error
	}
)
