package assets

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/honmaple/snow/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAssetsTestContext(t *testing.T, configure func(*core.Config)) *core.Context {
	t.Helper()

	conf := core.DefaultConfig()
	if configure != nil {
		configure(conf)
	}
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)
	return ctx
}

func TestNewReturnsErrorForMissingFiles(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.output", "css/app.css")
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hooks.assets.option.app.files is required")
}

func TestNewReturnsErrorForMissingOutput(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hooks.assets.option.app.output is required")
}

func TestNewReturnsErrorForUnknownFilter(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.filters", []string{"unknown"})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown assets filter "unknown"`)
}

func TestNewReturnsErrorForSassCompilerInFilters(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.filters", []string{"dartsass", "cssmin"})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown assets filter "dartsass"`)
}

func TestNewNormalizesStringFilters(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.filters", " cssmin, jsmin ")
	})

	hook, err := New(ctx)
	require.NoError(t, err)

	h := hook.(*AssetsHook)
	require.Contains(t, h.preAssetMap, "app")
	assert.Equal(t, sassCompilerLibSass, h.preAssetMap["app"].SassCompiler)
	assert.Equal(t, []string{"cssmin", "jsmin"}, h.preAssetMap["app"].Filters)
}

func TestNewUsesSassCompilerOption(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.sass_compiler", "dartsass")
		conf.Set("hooks.assets.option.app.filters", []string{"cssmin"})
	})

	hook, err := New(ctx)
	require.NoError(t, err)

	h := hook.(*AssetsHook)
	require.Contains(t, h.preAssetMap, "app")
	assert.Equal(t, sassCompilerDartSass, h.preAssetMap["app"].SassCompiler)
	assert.Equal(t, []string{"cssmin"}, h.preAssetMap["app"].Filters)
}

func TestNewReturnsErrorForUnknownSassCompiler(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.sass_compiler", "unknown")
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown Sass compiler "unknown"`)
}

func TestDartSassImportResolverLoadsPartial(t *testing.T) {
	assetsFS := fstest.MapFS{
		"scss/_variables.scss": &fstest.MapFile{Data: []byte("$color: red;")},
	}
	resolver := &dartSassImportResolver{
		assetsFS: assetsFS,
		dir:      "scss",
	}

	canonicalURL, err := resolver.CanonicalizeURL("variables")
	require.NoError(t, err)
	require.NotEmpty(t, canonicalURL)

	asset, err := resolver.Load(canonicalURL)
	require.NoError(t, err)
	assert.Equal(t, "$color: red;", asset.Content)
	assert.Equal(t, "SCSS", string(asset.SourceSyntax))
}

func TestDartSassImportResolverReturnsEmptyForMissingImport(t *testing.T) {
	resolver := &dartSassImportResolver{
		assetsFS: fstest.MapFS{},
		dir:      "scss",
	}

	canonicalURL, err := resolver.CanonicalizeURL("missing")
	require.NoError(t, err)
	assert.Empty(t, canonicalURL)
}

func TestDartSassImportResolverRejectsUnsupportedURL(t *testing.T) {
	resolver := &dartSassImportResolver{
		assetsFS: fstest.MapFS{},
		dir:      "scss",
	}

	_, err := resolver.Load("https://example.com/style.scss")
	require.Error(t, err)
	assert.ErrorIs(t, err, fs.ErrInvalid)
}
