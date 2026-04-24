package template

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
)

type scratch struct {
	m sync.Map
}

func (self *scratch) Set(key string, val any) any {
	self.m.Store(key, val)
	return val
}

func (self *scratch) Get(key string) any {
	val, ok := self.m.Load(key)
	if !ok {
		return nil
	}
	return val
}

func (self *scratch) Add(key string, newVal any) any {
	oldVal, ok := self.m.Load(key)
	if !ok {
		return nil
	}

	oval := pongo2.AsValue(oldVal)
	nval := pongo2.AsValue(newVal)
	if oval.IsNumber() && nval.IsNumber() {
		if oval.IsFloat() || nval.IsFloat() {
			newVal = oval.Float() + nval.Float()
		} else {
			newVal = oval.Integer() + nval.Integer()
		}
	} else if oval.IsString() && nval.IsString() {
		newVal = oval.String() + nval.String()
	} else {
		orval := reflect.ValueOf(oldVal)
		nrval := reflect.ValueOf(nval.Interface())
		switch orval.Kind() {
		case reflect.Array, reflect.Slice:
			switch nrval.Kind() {
			case reflect.Array, reflect.Slice:
				m := make([]any, orval.Len()+nrval.Len())
				for i := 0; i < orval.Len(); i++ {
					m[i] = orval.Index(i).Interface()
				}
				oi := orval.Len()
				for i := 0; i < nrval.Len(); i++ {
					m[i+oi] = nrval.Index(i).Interface()
				}
				newVal = m
			default:
				m := make([]any, orval.Len()+1)
				for i := 0; i < orval.Len(); i++ {
					m[i] = orval.Index(i).Interface()
				}
				m[orval.Len()] = nrval.Interface()
				newVal = m
			}
		}
	}
	self.m.Store(key, newVal)
	return newVal
}

func (self *scratch) JSON(key string) any {
	val, ok := self.m.Load(key)
	if !ok {
		return nil
	}
	buf, err := json.Marshal(val)
	if err == nil {
		return string(buf)
	}
	return err.Error()
}

func newScratch(ctx *core.Context, vars map[string]any) any {
	return &scratch{}
}

func newScratchFunc(ctx *core.Context, vars map[string]any) any {
	return func() any {
		return &scratch{}
	}
}

func init() {
	Register("scratch", func(ctx *core.Context, set TemplateSet) error {
		set.Register("scratch", newScratch)
		set.Register("newScratch", newScratchFunc)
		return nil
	})
}
