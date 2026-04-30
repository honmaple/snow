package encrypt

import (
	"fmt"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
)

type (
	Option struct {
		Password    string `json:"password"`
		Description string `json:"description"`
	}
	EncryptHook struct {
		hook.HookImpl
		ctx *core.Context
		opt Option
	}
)

func (h *EncryptHook) HandlePage(page *content.Page) *content.Page {
	password := page.FrontMatter.GetString("password")
	if password == "" {
		return page
	}
	description := h.opt.Description
	if v := strings.SplitN(password, ",", 2); len(v) == 2 {
		password = v[0]
		description = v[1]
	}
	if description == "" {
		description = "这是一篇加密的文章，你需要输入正确的密码."
	}
	page.Summary = fmt.Sprintf(`<shortcode _name="encrypt" password="%s" description="%s">%s</shortcode>`, password, description, page.Summary)
	page.Content = fmt.Sprintf(`<shortcode _name="encrypt" password="%s" description="%s">%s</shortcode>`, password, description, page.Content)
	return page
}

func New(ctx *core.Context) (hook.Hook, error) {
	var opt Option
	if err := hook.Unmarshal(ctx.Config.Get("hooks.encrypt.option"), &opt); err != nil {
		return nil, err
	}

	e := &EncryptHook{
		ctx: ctx,
		opt: opt,
	}
	return e, nil
}

func init() {
	hook.Register("encrypt", New)
}
