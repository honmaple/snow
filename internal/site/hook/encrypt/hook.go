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
		Mode        string `json:"mode"`
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
	modeAttr := ""
	if h.opt.Mode != "" {
		modeAttr = fmt.Sprintf(` mode="%s"`, h.opt.Mode)
	}
	page.Summary = fmt.Sprintf(`<shortcode encrypt password="%s" description="%s"%s>%s</shortcode>`, password, description, modeAttr, page.Summary)
	page.Content = fmt.Sprintf(`<shortcode encrypt password="%s" description="%s"%s>%s</shortcode>`, password, description, modeAttr, page.Content)
	return page
}

func New(ctx *core.Context) (hook.Hook, error) {
	var opt Option
	if err := hook.Unmarshal(ctx.Config.Get("hooks.encrypt.option"), &opt); err != nil {
		return nil, err
	}
	mode, err := normalizeMode(opt.Mode)
	if err != nil {
		return nil, err
	}
	opt.Mode = mode

	e := &EncryptHook{
		ctx: ctx,
		opt: opt,
	}
	return e, nil
}

func init() {
	hook.Register("encrypt", New)
}
