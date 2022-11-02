package hook

import (
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/static/hook/webassets"
)

func init() {
	hook.Register("webassets", webassets.New)
}
