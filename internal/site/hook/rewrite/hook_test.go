package rewrite

import (
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRewriteTestContext(configure func(*core.Config)) *core.Context {
	conf := core.NewConfig()
	if configure != nil {
		configure(conf)
	}
	return &core.Context{
		LocaleContext: &core.LocaleContext{Config: conf},
	}
}

func TestNewReturnsErrorForMissingSrc(t *testing.T) {
	ctx := newRewriteTestContext(func(conf *core.Config) {
		conf.Set("hooks.rewrite.option", []map[string]any{
			{"dst": "tags"},
		})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hooks.rewrite.option[0].src is required")
}

func TestNewReturnsErrorForMissingDst(t *testing.T) {
	ctx := newRewriteTestContext(func(conf *core.Config) {
		conf.Set("hooks.rewrite.option", []map[string]any{
			{"src": "tag"},
		})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hooks.rewrite.option[0].dst is required")
}

func TestNewReturnsErrorForInvalidType(t *testing.T) {
	ctx := newRewriteTestContext(func(conf *core.Config) {
		conf.Set("hooks.rewrite.option", []map[string]any{
			{"src": "tag", "dst": "tags", "type": "unknown"},
		})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `hooks.rewrite.option[0].type "unknown" is invalid`)
}

func TestNewAcceptsEmptyAndListType(t *testing.T) {
	ctx := newRewriteTestContext(func(conf *core.Config) {
		conf.Set("hooks.rewrite.option", []map[string]any{
			{"src": "tag", "dst": "tags"},
			{"src": "category", "dst": "categories", "type": "list"},
		})
	})

	hook, err := New(ctx)
	require.NoError(t, err)

	h := hook.(*RewriteHook)
	require.Len(t, h.opts, 2)
	assert.Equal(t, "list", h.opts[1].Type)
}
