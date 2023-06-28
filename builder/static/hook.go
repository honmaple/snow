package static

type (
	Hook interface {
		Static(*Static) *Static
		Statics(Statics) Statics
	}
	Hooks []Hook
)

func (hooks Hooks) Static(static *Static) *Static {
	for _, hook := range hooks {
		static = hook.Static(static)
		if static == nil {
			return nil
		}
	}
	return static
}

func (hooks Hooks) Statics(statics Statics) Statics {
	for _, hook := range hooks {
		statics = hook.Statics(statics)
		if len(statics) == 0 {
			return nil
		}
	}
	return statics
}
