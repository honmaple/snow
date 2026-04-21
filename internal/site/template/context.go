package template

type (
	Context interface {
		GetPrivate(string) any
		SetPrivate(string, any)
	}
	contextImpl struct {
		m map[string]any
	}
)

func (ctx *contextImpl) Get(k string) any {
	return ctx.m[k]
}

func (ctx *contextImpl) Set(k string, v any) {
	ctx.m[k] = v
}

func (ctx *contextImpl) SetPrivate(k string, v any) {
	val, ok := ctx.m["__snow__"].(*contextImpl)
	if !ok {
		val = &contextImpl{m: make(map[string]any)}
	}
	val.m[k] = v
	ctx.m["__snow__"] = val
}

func (ctx *contextImpl) GetPrivate(k string) any {
	snow, ok := ctx.m["__snow__"].(*contextImpl)
	if !ok {
		return nil
	}
	return snow.m[k]
}

func SetPrivate(ctx map[string]any, key string, value any) {
	val, ok := ctx["__snow__"].(*contextImpl)
	if !ok {
		val = &contextImpl{m: make(map[string]any)}
	}
	val.m[key] = value
	ctx["__snow__"] = val
}

func GetPrivate[T any](ctx map[string]any, key string) (result T, ok bool) {
	snow, ok := ctx["__snow__"].(*contextImpl)
	if !ok {
		return result, false
	}
	result, ok = snow.m[key].(T)
	return
}

func NewContext(m map[string]any) Context {
	return &contextImpl{m: m}

}
