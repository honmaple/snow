package snakecase

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/iancoleman/strcase"
)

type Template struct {
	template.Template
}

type visit struct {
	typ reflect.Type
	ptr uintptr
}

func (tpl *Template) toSnakeCase(str string) string {
	return strcase.ToSnake(str)
}

func (tpl *Template) convertArg(arg *pongo2.Value, targetType reflect.Type) (reflect.Value, error) {
	if targetType == reflect.TypeOf((*pongo2.Value)(nil)) {
		return reflect.ValueOf(arg), nil
	}

	if arg == nil || arg.Interface() == nil {
		switch targetType.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			return reflect.Zero(targetType), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot use nil as %s", targetType)
		}
	}

	var goArg any
	switch targetType.Kind() {
	case reflect.Bool:
		goArg = arg.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		goArg = arg.Integer()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		goArg = uint(arg.Integer())
	case reflect.Float32, reflect.Float64:
		goArg = arg.Float()
	case reflect.String:
		goArg = arg.String()
	default:
		goArg = arg.Interface()
	}

	val := reflect.ValueOf(goArg)
	if !val.IsValid() {
		return reflect.Zero(targetType), nil
	}
	if val.Type().AssignableTo(targetType) {
		return val, nil
	}
	if val.Type().ConvertibleTo(targetType) {
		return val.Convert(targetType), nil
	}
	return reflect.Value{}, fmt.Errorf("cannot use %s as %s", val.Type(), targetType)
}

func (tpl *Template) callFunc(fn reflect.Value, args ...*pongo2.Value) (out *pongo2.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			out = nil
			err = fmt.Errorf("call function: %v", r)
		}
	}()

	fnType := fn.Type()
	numIn := fnType.NumIn()
	if fnType.IsVariadic() {
		if len(args) < numIn-1 {
			return nil, fmt.Errorf("function expects at least %d arguments, got %d", numIn-1, len(args))
		}
	} else if len(args) != numIn {
		return nil, fmt.Errorf("function expects %d arguments, got %d", numIn, len(args))
	}

	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		var targetType reflect.Type
		if fnType.IsVariadic() && i >= numIn-1 {
			targetType = fnType.In(numIn - 1).Elem()
		} else {
			targetType = fnType.In(i)
		}

		reflectArg, err := tpl.convertArg(arg, targetType)
		if err != nil {
			return nil, err
		}
		reflectArgs[i] = reflectArg
	}

	results := fn.Call(reflectArgs)
	if len(results) == 0 {
		return pongo2.AsValue(""), nil
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	last := results[len(results)-1]
	hasError := last.IsValid() && last.Type().Implements(errorType)
	if hasError {
		switch last.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			if last.IsNil() {
				break
			}
			return nil, last.Interface().(error)
		default:
			return nil, last.Interface().(error)
		}
	}

	switch {
	case len(results) == 1 && hasError:
		return pongo2.AsValue(""), nil
	case len(results) == 1:
		return pongo2.AsValue(tpl.wrapValue(results[0].Interface())), nil
	case len(results) == 2 && hasError:
		return pongo2.AsValue(tpl.wrapValue(results[0].Interface())), nil
	default:
		return nil, errors.New("function must return value, value/error, error, or nothing")
	}
}

func (tpl *Template) wrapFunc(fn reflect.Value) any {
	return func(args ...*pongo2.Value) (*pongo2.Value, error) {
		return tpl.callFunc(fn, args...)
	}
}

func (tpl *Template) toSnakeMap(obj any, seen map[visit]any) map[string]any {
	snakeMap := make(map[string]any)
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)
	if val.Kind() == reflect.Ptr && !val.IsNil() && val.Elem().Kind() == reflect.Struct {
		key := visit{typ: typ, ptr: val.Pointer()}
		if cached, ok := seen[key]; ok {
			return cached.(map[string]any)
		}
		seen[key] = snakeMap
	}

	for i := 0; i < val.NumMethod(); i++ {
		methodType := typ.Method(i)
		methodVal := val.Method(i)
		snakeMethodName := tpl.toSnakeCase(methodType.Name)

		localMethodVal := methodVal
		snakeMap[snakeMethodName] = func(args ...*pongo2.Value) (*pongo2.Value, error) {
			return tpl.callFunc(localMethodVal, args...)
		}
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		t := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			// 略过私有字段
			if !field.CanInterface() {
				continue
			}
			snakeFieldName := tpl.toSnakeCase(t.Field(i).Name)
			// 递归转换字段的值
			snakeMap[snakeFieldName] = tpl.wrapValueSeen(field.Interface(), seen)
		}
	}
	return snakeMap
}

func (tpl *Template) wrapValue(v any) any {
	return tpl.wrapValueSeen(v, make(map[visit]any))
}

func (tpl *Template) wrapValueSeen(v any, seen map[visit]any) any {
	if v == nil {
		return nil
	}
	if _, ok := v.(time.Time); ok {
		return v
	}

	val := reflect.ValueOf(v)
	kind := val.Kind()

	if kind == reflect.Func {
		if val.IsNil() {
			return v
		}
		return tpl.wrapFunc(val)
	}

	if kind == reflect.Ptr {
		if val.IsNil() {
			return v
		}
		if val.Elem().Kind() == reflect.Struct {
			return tpl.toSnakeMap(v, seen)
		}
		kind = val.Elem().Kind()
	}

	if kind == reflect.Struct {
		return tpl.toSnakeMap(v, seen)
	}

	if kind == reflect.Map {
		if pCtx, ok := v.(pongo2.Context); ok {
			return tpl.wrapContextSeen(pCtx, seen)
		}
		if gMap, ok := v.(map[string]any); ok {
			newMap := make(map[string]any)
			for mk, mv := range gMap {
				newMap[mk] = tpl.wrapValueSeen(mv, seen)
			}
			return newMap
		}
		if val.Type().Key().Kind() == reflect.String {
			newMap := make(map[string]any)
			iter := val.MapRange()
			for iter.Next() {
				newMap[iter.Key().String()] = tpl.wrapValueSeen(iter.Value().Interface(), seen)
			}
			return newMap
		}
	}

	if kind == reflect.Slice || kind == reflect.Array {
		length := val.Len()
		newSlice := make([]any, length)
		for i := 0; i < length; i++ {
			newSlice[i] = tpl.wrapValueSeen(val.Index(i).Interface(), seen)
		}
		return newSlice
	}
	return v
}

func (tpl *Template) wrapContext(ctx map[string]any) map[string]any {
	return tpl.wrapContextSeen(ctx, make(map[visit]any))
}

func (tpl *Template) wrapContextSeen(ctx map[string]any, seen map[visit]any) map[string]any {
	if ctx == nil {
		return nil
	}
	newCtx := make(map[string]any)
	for k, v := range ctx {
		newCtx[k] = tpl.wrapValueSeen(v, seen)
	}
	return newCtx
}

func (tpl *Template) Execute(ctx map[string]any) (string, error) {
	return tpl.Template.Execute(tpl.wrapContext(ctx))
}

type TemplateSet struct {
	template.TemplateSet
}

func (set *TemplateSet) newTemplate(tpl template.Template) template.Template {
	return &Template{Template: tpl}
}

func (set *TemplateSet) Lookup(names ...string) template.Template {
	tpl := set.TemplateSet.Lookup(names...)
	if tpl == nil {
		return nil
	}
	return set.newTemplate(tpl)
}

func (set *TemplateSet) FromFile(name string) (template.Template, error) {
	tpl, err := set.TemplateSet.FromFile(name)
	if err != nil {
		return nil, err
	}
	return set.newTemplate(tpl), nil
}

func (set *TemplateSet) FromBytes(b []byte) (template.Template, error) {
	tpl, err := set.TemplateSet.FromBytes(b)
	if err != nil {
		return nil, err
	}
	return set.newTemplate(tpl), nil
}

func (set *TemplateSet) FromString(b string) (template.Template, error) {
	tpl, err := set.TemplateSet.FromString(b)
	if err != nil {
		return nil, err
	}
	return set.newTemplate(tpl), nil
}
