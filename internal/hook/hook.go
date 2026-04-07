package hook

import (
	"github.com/honmaple/snow/internal/content"
	"github.com/honmaple/snow/internal/static"
)

type (
	StaticHook       = static.Hook
	ContentHook      = content.Hook
	StaticEmptyHook  = static.EmptyHook
	ContentEmptyHook = content.EmptyHook

	BuildHook interface {
		AfterBuild() error
		BeforeBuild() error
	}
	Hook interface {
		BuildHook
		StaticHook
		ContentHook
	}
)

type HookImpl struct {
	StaticEmptyHook
	ContentEmptyHook
}

func (HookImpl) AfterBuild() error  { return nil }
func (HookImpl) BeforeBuild() error { return nil }
