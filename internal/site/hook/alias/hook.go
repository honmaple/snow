package alias

import (
	"context"
	"fmt"
	"html"
	stdpath "path"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
)

type AliasHook struct {
	hook.HookImpl

	ctx     *core.Context
	aliases []Alias
}

type Alias struct {
	From string
	To   string
}

func aliasFile(path string) string {
	if strings.HasSuffix(path, "/") {
		return path + "index.html"
	}
	if stdpath.Ext(path) == "" {
		return path + "/index.html"
	}
	return path
}

func aliasHTML(target string) string {
	target = html.EscapeString(target)
	return fmt.Sprintf(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="refresh" content="0; url=%s">
<link rel="canonical" href="%s">
</head>
<body>
<a href="%s">Redirect</a>
</body>
</html>
`, target, target, target)
}

func parseAliases(values []string) ([]Alias, error) {
	aliases := make([]Alias, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid alias %q: expected old:new", value)
		}
		from, err := normalizeAliasPath(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid alias %q: %w", value, err)
		}
		alias := Alias{
			From: from,
			To:   strings.TrimSpace(parts[1]),
		}
		if alias.From == "" || alias.To == "" {
			return nil, fmt.Errorf("invalid alias %q: old and new URL are required", value)
		}
		aliases = append(aliases, alias)
	}
	return aliases, nil
}

func normalizeAliasPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if strings.ContainsAny(path, "?#\\") {
		return "", fmt.Errorf("old URL %q contains invalid path characters", path)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	file := aliasFile(path)
	if file == "" || stdpath.Clean(file) != file {
		return "", fmt.Errorf("old URL %q is not a clean output path", path)
	}
	return path, nil
}

func (h *AliasHook) AfterBuild(ctx context.Context, writer core.Writer) error {
	for _, alias := range h.aliases {
		if h.ctx != nil && h.ctx.Logger != nil {
			h.ctx.Logger.Debugf("write alias [%s] -> %s", alias.From, alias.To)
		}
		if err := writer.WriteFile(ctx, aliasFile(alias.From), strings.NewReader(aliasHTML(h.ctx.GetURL(alias.To)))); err != nil {
			return err
		}
	}
	return nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	aliases, err := parseAliases(ctx.Config.GetStringSlice("hooks.alias.option"))
	if err != nil {
		return nil, err
	}
	return &AliasHook{ctx: ctx, aliases: aliases}, nil
}

func init() {
	hook.Register("alias", New)
}
