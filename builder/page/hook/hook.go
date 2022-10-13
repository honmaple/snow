package hook

import (
	"github.com/honmaple/snow/builder/hook"
)

func init() {
	hook.Register("feed", newFeed)
	hook.Register("encrypt", newEncrypy)
}
