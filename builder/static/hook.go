package static

type (
	Hook interface {
		BeforeStaticWrite(*Static) *Static
		BeforeStaticsWrite(Statics) Statics
	}
	Hooks []Hook
)

func (hooks Hooks) BeforeStaticWrite(static *Static) *Static {
	for _, hook := range hooks {
		static = hook.BeforeStaticWrite(static)
	}
	return static
}

func (hooks Hooks) BeforeStaticsWrite(statics Statics) Statics {
	for _, hook := range hooks {
		statics = hook.BeforeStaticsWrite(statics)
	}
	return statics
}
