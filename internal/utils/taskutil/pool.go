package taskutil

import (
	"github.com/alitto/pond/v2"
)

type Pool[A any] interface {
	Invoke(A) pond.Task
	Submit(func() error) pond.Task
	StopAndWait()
}

type defaultPool[A any] struct {
	fn   func(A) error
	pool pond.Pool
}

func (p *defaultPool[A]) Invoke(arg A) pond.Task {
	return p.pool.SubmitErr(func() error {
		return p.fn(arg)
	})
}

func (p *defaultPool[A]) Submit(fn func() error) pond.Task {
	return p.pool.SubmitErr(fn)
}

func (p *defaultPool[A]) StopAndWait() {
	p.pool.StopAndWait()
}

func NewPool[A any](size int, fn func(arg A) error) Pool[A] {
	p := pond.NewPool(size)
	return &defaultPool[A]{fn: fn, pool: p}
}

type ResultPool[A any, R any] interface {
	Invoke(A) pond.ResultTask[R]
	Submit(func() (R, error)) pond.ResultTask[R]
	StopAndWait()
}

type resultPool[A any, R any] struct {
	fn   func(A) (R, error)
	pool pond.ResultPool[R]
}

func (p *resultPool[A, R]) Invoke(arg A) pond.ResultTask[R] {
	return p.pool.SubmitErr(func() (R, error) {
		return p.fn(arg)
	})
}

func (p *resultPool[A, R]) Submit(fn func() (R, error)) pond.ResultTask[R] {
	return p.pool.SubmitErr(fn)
}

func (p *resultPool[A, R]) StopAndWait() {
	p.pool.StopAndWait()
}

func NewResultPool[A any, R any](size int, fn func(arg A) (R, error)) ResultPool[A, R] {
	p := pond.NewResultPool[R](size)
	return &resultPool[A, R]{fn: fn, pool: p}
}
