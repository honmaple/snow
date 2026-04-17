package assets

import (
	"fmt"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/hook"
	"github.com/honmaple/snow/internal/theme/template"
)

type (
	assets struct {
		hook.HookImpl
		ctx       *core.Context
		collector *AssetsCollector
	}
)

func (h *assets) BeforeBuild() error {
	return nil
}

// 写入收集的文件
func (h *assets) AfterBuild() error {
	for _, asset := range h.collector.assets {
		fmt.Println(asset.Output)
	}
	return nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	h := &assets{
		ctx:       ctx,
		collector: NewAssetsCollector(ctx),
	}

	template.RegisterTransient("__assets_collector", h.collector)
	return h, nil
}

func init() {
	hook.Register("assets", New)
}
