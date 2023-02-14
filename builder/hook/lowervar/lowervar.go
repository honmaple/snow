package lowervar

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type Lowervar struct {
	hook.BaseHook
	conf config.Config
}

func toLower(val reflect.Value, tagName string) interface{} {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		if val.Len() == 0 {
			return nil
		}
		m := make(map[string]interface{})
		for _, k := range val.MapKeys() {
			m[k.String()] = toLower(val.MapIndex(k), tagName)
		}
		return m
	case reflect.Array, reflect.Slice:
		if val.Len() == 0 {
			return nil
		}
		m := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			m[i] = toLower(val.Index(i), tagName)
		}
		return m
	case reflect.Struct:
		t := val.Type()
		m := make(map[string]interface{})
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tag := field.Tag.Get(tagName)
			if tag == "-" {
				continue
			}

			inline := false
			if field.Anonymous {
				inline = true
			} else if !field.IsExported() {
				continue
			} else {
				opts := strings.Split(tag, ",")
				if len(opts) > 1 {
					for _, flag := range opts[1:] {
						switch flag {
						case "inline":
							inline = true
						}
					}
					tag = opts[0]
				}
			}
			if tag == "" {
				tag = field.Name
			}

			vfield := val.Field(i)
			if inline {
				kind := field.Type.Kind()
				if kind == reflect.Ptr {
					if vfield.IsNil() {
						continue
					}
					vfield = vfield.Elem()
				}
				if !field.IsExported() && kind != reflect.Struct {
					continue
				}

				inlineVal := toLower(vfield, tagName)
				if inlineVal == nil {
					continue
				}
				for k, v := range inlineVal.(map[string]interface{}) {
					m[k] = v
				}
				continue
			}
			m[tag] = toLower(vfield, tagName)
		}
		return m
	// case reflect.Interface:
	//	fmt.Println(val.Addr().Kind())
	//	return ""
	default:
		fmt.Println(val.Interface(), val.Kind())
		return val.Interface()
	}
}

func (e *Lowervar) BeforeWrite(vars map[string]interface{}) {
	// d, _ := mapstructure.NewDecoder(mapstructure.DecoderConfig{TagName: "json"})
	for k, v := range vars {
		vars[k] = toLower(reflect.ValueOf(v), "json")
	}
}

func (e *Lowervar) Name() string {
	return "lowervar"
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	e := &Lowervar{
		conf: conf,
	}

	return e
}

func init() {
	hook.Register("lowervar", New)
}
