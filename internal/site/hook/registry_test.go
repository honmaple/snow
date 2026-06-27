package hook

import (
	"bytes"
	"strings"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withFactories(t *testing.T, m map[string]Factory) {
	t.Helper()

	old := factories
	factories = m
	t.Cleanup(func() {
		factories = old
	})
}

func newTestContext(conf *core.Config, logger core.Logger) *core.Context {
	if conf == nil {
		conf = core.NewConfig()
	}
	if logger == nil {
		logger = logrus.New()
	}
	return &core.Context{
		LocaleContext: &core.LocaleContext{
			Config: conf,
		},
		Logger:         logger,
		OtherLanguages: make(map[string]*core.LocaleContext),
	}
}

func noopFactory(*core.Context) (Hook, error) {
	return HookImpl{}, nil
}

func TestNewReturnsErrorForEnabledUnregisteredHook(t *testing.T) {
	withFactories(t, map[string]Factory{})

	conf := core.NewConfig()
	conf.Set("hooks.missing.enabled", true)
	ctx := newTestContext(conf, nil)

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `hook "missing" is enabled but not registered`)
}

func TestNewIgnoresDisabledUnregisteredHook(t *testing.T) {
	withFactories(t, map[string]Factory{})

	conf := core.NewConfig()
	conf.Set("hooks.missing.weight", 10)
	conf.Set("hooks.missing.enabled", false)
	ctx := newTestContext(conf, nil)

	registry, err := New(ctx)
	require.NoError(t, err)
	assert.Empty(t, registry.names)
	assert.Empty(t, registry.hooks)
}

func TestNewUsesDefaultHookOrder(t *testing.T) {
	withFactories(t, map[string]Factory{
		"assets":    noopFactory,
		"encrypt":   noopFactory,
		"links":     noopFactory,
		"shortcode": noopFactory,
	})

	ctx := newTestContext(core.DefaultConfig(), nil)
	registry, err := New(ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"assets", "encrypt", "links", "shortcode"}, registry.names)
}

func TestNewSortsHooksByWeightThenName(t *testing.T) {
	withFactories(t, map[string]Factory{
		"rewrite": noopFactory,
		"pelican": noopFactory,
	})

	conf := core.NewConfig()
	conf.Set("hooks.rewrite.enabled", true)
	conf.Set("hooks.rewrite.weight", 30)
	conf.Set("hooks.pelican.enabled", true)
	conf.Set("hooks.pelican.weight", 30)
	ctx := newTestContext(conf, nil)

	registry, err := New(ctx)
	require.NoError(t, err)

	assert.Equal(t, []string{"pelican", "rewrite"}, registry.names)
}

func TestNewLogsEnabledHookOrderAtDebugLevel(t *testing.T) {
	withFactories(t, map[string]Factory{
		"assets":    noopFactory,
		"encrypt":   noopFactory,
		"links":     noopFactory,
		"shortcode": noopFactory,
	})

	var buf bytes.Buffer
	logger := logrus.New()
	logger.Out = &buf
	logger.Level = logrus.DebugLevel
	logger.Formatter = &logrus.TextFormatter{DisableTimestamp: true}

	ctx := newTestContext(core.DefaultConfig(), logger)
	_, err := New(ctx)
	require.NoError(t, err)

	assert.Contains(t, strings.TrimSpace(buf.String()), "Enabled hooks: assets(20), encrypt(50), links(55), shortcode(60)")
}
