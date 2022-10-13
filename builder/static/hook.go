package static

type (
	Hook interface {
		BeforeStatic(*Static) *Static
		BeforeStaticList(Statics) Statics
	}
	Hooks []Hook
)

func (hooks Hooks) BeforeStatic(static *Static) *Static {
	for _, hook := range hooks {
		static = hook.BeforeStatic(static)
	}
	return static
}

func (hooks Hooks) BeforeStaticList(statics Statics) Statics {
	for _, hook := range hooks {
		statics = hook.BeforeStaticList(statics)
	}
	return statics
}
