package filter

import (
	"testing"

	"github.com/expr-lang/expr"
	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newFilterTestContext(expr string) *core.Context {
	conf := core.NewConfig()
	if expr != "" {
		conf.Set("hooks.filter.option.page_filter", expr)
	}
	return &core.Context{
		LocaleContext: &core.LocaleContext{Config: conf},
	}
}

func TestNewReturnsErrorForInvalidPageFilter(t *testing.T) {
	ctx := newFilterTestContext("title ==")

	_, err := New(ctx)
	require.Error(t, err)
}

func TestHandlePageUsesCompiledPageFilter(t *testing.T) {
	ctx := newFilterTestContext(`title == "keep"`)

	hook, err := New(ctx)
	require.NoError(t, err)

	h := hook.(*FilterHook)
	assert.NotNil(t, h.tpl)

	keep := &content.Page{
		Node: &content.Node{
			FrontMatter: content.NewFrontMatter(map[string]any{"title": "keep"}),
		},
	}
	drop := &content.Page{
		Node: &content.Node{
			FrontMatter: content.NewFrontMatter(map[string]any{"title": "drop"}),
		},
	}

	assert.Same(t, keep, h.HandlePage(keep))
	assert.Nil(t, h.HandlePage(drop))
}

// BenchmarkFilterWithExpr
// BenchmarkFilterWithExpr-10		  245794		  4891 ns/op
// BenchmarkFilterWithPongo2
// BenchmarkFilterWithPongo2-10		  862621		  1277 ns/op
func BenchmarkFilterWithExpr(b *testing.B) {
	env := map[string]any{
		"tags": []string{"tag1", "tag2", "tag3"},
	}
	for b.Loop() {
		program, err := expr.Compile(`'tag2' in tags`, expr.Env(env))
		if err != nil {
			panic(err)
		}
		_, err = expr.Run(program, env)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFilterWithPongo2(b *testing.B) {
	ctx := map[string]any{
		"tags": []string{"tag1", "tag2", "tag3"},
	}
	for b.Loop() {
		tpl, err := pongo2.FromString("{{ 'tag2' in tags }}")
		if err != nil {
			panic(err)
		}
		_, err = tpl.Execute(ctx)
		if err != nil {
			panic(err)
		}
	}
}
