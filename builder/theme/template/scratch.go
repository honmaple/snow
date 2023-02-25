package template

import (
	"reflect"
	"sync"

	"encoding/json"

	"github.com/flosch/pongo2/v6"
)

type scratch struct {
	m sync.Map
}

func (self *scratch) Set(key string, val interface{}) interface{} {
	self.m.Store(key, val)
	return val
}

func (self *scratch) Get(key string) interface{} {
	val, ok := self.m.Load(key)
	if !ok {
		return nil
	}
	return val
}

func (self *scratch) Add(key string, newVal interface{}) interface{} {
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
				m := make([]interface{}, orval.Len()+nrval.Len())
				for i := 0; i < orval.Len(); i++ {
					m[i] = orval.Index(i).Interface()
				}
				oi := orval.Len()
				for i := 0; i < nrval.Len(); i++ {
					m[i+oi] = nrval.Index(i).Interface()
				}
				newVal = m
			default:
				m := make([]interface{}, orval.Len()+1)
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

func (self *scratch) JSON(key string) interface{} {
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

func newScratch(ctx map[string]interface{}) interface{} {
	return &scratch{}
}

func newScratchFunc(ctx map[string]interface{}) interface{} {
	return func() interface{} {
		return &scratch{}
	}
}
