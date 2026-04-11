package hook

import (
	"github.com/honmaple/snow/internal/content"
)

type (
	ContentHook      = content.Hook
	ContentEmptyHook = content.EmptyHook

	BuildHook interface {
		AfterBuild() error
		BeforeBuild() error
	}
	Hook interface {
		BuildHook
		ContentHook
	}
)

type HookImpl struct {
	ContentEmptyHook
}

func (HookImpl) AfterBuild() error  { return nil }
func (HookImpl) BeforeBuild() error { return nil }
