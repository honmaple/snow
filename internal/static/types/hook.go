package types

type (
	Hook interface {
		HandleStatic(*Static) *Static
		HandleStatics(Statics) Statics
	}
	EmptyHook struct {
	}
)

func (EmptyHook) HandleStatic(result *Static) *Static   { return result }
func (EmptyHook) HandleStatics(results Statics) Statics { return results }
