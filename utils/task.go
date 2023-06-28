package utils

import (
	"sync"

	"github.com/panjf2000/ants/v2"
)

type taskPool struct {
	*ants.PoolWithFunc
	wg *sync.WaitGroup
}

func (p *taskPool) Invoke(i interface{}) {
	p.wg.Add(1)
	p.PoolWithFunc.Invoke(i)
}

func (p *taskPool) Wait() {
	p.wg.Wait()
}

func NewTaskPool(wg *sync.WaitGroup, size int, f func(interface{})) *taskPool {
	p, _ := ants.NewPoolWithFunc(size, f)
	return &taskPool{
		PoolWithFunc: p,
		wg:           wg,
	}
}
