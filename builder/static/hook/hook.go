package hook

import (
	"github.com/honmaple/snow/builder/hook"
)

func init() {
	hook.Register("minify", newMinify)
}
