package assets

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/writer"
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

func filterNames(filters []Filter) []string {
	names := make([]string, 0, len(filters))
	for _, filter := range filters {
		names = append(names, filter.Name())
	}
	return names
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
	assert.Equal(t, sassCompilerLibSass, h.preAssetMap["app"].SassCompiler.Name())
	assert.Equal(t, []string{assetFilterCSSMin, assetFilterJSMin}, filterNames(h.preAssetMap["app"].Filters))
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
	assert.Equal(t, sassCompilerDartSass, h.preAssetMap["app"].SassCompiler.Name())
	assert.Equal(t, []string{assetFilterCSSMin}, filterNames(h.preAssetMap["app"].Filters))
}

func TestNewReturnsErrorForUnknownSassCompiler(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.sass_compiler", "unknown")
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown sass compiler "unknown"`)
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

func TestNewNormalizesImageFilterOptions(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.cover.files", []string{"images/cover.jpg"})
		conf.Set("hooks.assets.option.cover.output", "images/cover-small.jpg")
		conf.Set("hooks.assets.option.cover.filters", []any{
			map[string]any{
				"name":    "image",
				"width":   1200,
				"height":  800,
				"fit":     "cover",
				"quality": 82,
			},
		})
	})

	hook, err := New(ctx)
	require.NoError(t, err)

	h := hook.(*AssetsHook)
	require.Contains(t, h.preAssetMap, "cover")
	require.Len(t, h.preAssetMap["cover"].Filters, 1)
	imageFilter := h.preAssetMap["cover"].Filters[0].(*ImageFilter)
	assert.Equal(t, assetFilterImage, imageFilter.Name())
	assert.Equal(t, 1200, imageFilter.Width)
	assert.Equal(t, 800, imageFilter.Height)
	assert.Equal(t, imageFitCover, imageFilter.Fit)
	assert.Equal(t, 82, imageFilter.Quality)
	assert.Equal(t, "jpg", imageFilter.Format)
}

func TestNormalizeFiltersSupportsTypedMapSlice(t *testing.T) {
	filters, err := normalizeFilters([]map[string]any{
		{"name": "image", "width": 320, "format": "jpg"},
	})
	require.NoError(t, err)

	require.Len(t, filters, 1)
	imageFilter := filters[0].(*ImageFilter)
	assert.Equal(t, assetFilterImage, imageFilter.Name())
	assert.Equal(t, 320, imageFilter.Width)
	assert.Equal(t, imageFitInside, imageFilter.Fit)
	assert.Equal(t, 85, imageFilter.Quality)
	assert.Equal(t, "jpg", imageFilter.Format)
}

func TestWithImageFormatOptionAddsOutputFormat(t *testing.T) {
	filters, err := normalizeFilters(withImageFormatOption([]map[string]any{
		{"name": "image", "width": 320},
	}, "cover.png"))
	require.NoError(t, err)

	require.Len(t, filters, 1)
	imageFilter := filters[0].(*ImageFilter)
	assert.Equal(t, "png", imageFilter.Format)
}

func TestImageFilterAppliesAndEncodesWithoutAsset(t *testing.T) {
	filter, err := newImageFilter(map[string]any{
		"width":   25,
		"quality": 80,
		"format":  "jpg",
	})
	require.NoError(t, err)

	img := filter.Apply(testImage(100, 80))
	assert.Equal(t, image.Rect(0, 0, 25, 20), img.Bounds())

	var buf bytes.Buffer
	require.NoError(t, filter.Encode(&buf, img, "jpg"))

	decoded, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 25, 20), decoded.Bounds())
}

func TestNewReturnsErrorForUnknownImageFilterOption(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.cover.files", []string{"images/cover.jpg"})
		conf.Set("hooks.assets.option.cover.output", "images/cover-small.jpg")
		conf.Set("hooks.assets.option.cover.filters", []any{
			map[string]any{"name": "image", "unknown": true},
		})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown image filter option "unknown"`)
}

func TestNewReturnsErrorForInvalidImageQuality(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.cover.files", []string{"images/cover.jpg"})
		conf.Set("hooks.assets.option.cover.output", "images/cover-small.jpg")
		conf.Set("hooks.assets.option.cover.filters", []any{
			map[string]any{"name": "image", "quality": 101},
		})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image filter quality must be between 1 and 100")
}

func TestNewReturnsErrorForFilterOptionsOnCSSMin(t *testing.T) {
	ctx := newAssetsTestContext(t, func(conf *core.Config) {
		conf.Set("hooks.assets.option.app.files", []string{"css/app.css"})
		conf.Set("hooks.assets.option.app.output", "css/app.css")
		conf.Set("hooks.assets.option.app.filters", []any{
			map[string]any{"name": "cssmin", "quality": 80},
		})
	})

	_, err := New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `assets filter "cssmin" does not accept options`)
}

func TestAssetExecuteResizesJPEGImage(t *testing.T) {
	asset := &Asset{
		Files:  []string{"images/cover.jpg"},
		Output: "cover-small.jpg",
		Filters: []Filter{
			&ImageFilter{Width: 50, Height: 50, Fit: imageFitInside, Quality: 80, Format: "jpg"},
		},
	}
	w := writer.NewMemoryWriter()

	err := asset.Execute(context.Background(), fstest.MapFS{
		"images/cover.jpg": &fstest.MapFile{Data: testJPEG(t, 100, 80)},
	}, w)
	require.NoError(t, err)

	img := decodeOutputImage(t, w, "cover-small.jpg")
	assert.Equal(t, image.Rect(0, 0, 50, 40), img.Bounds())
}

func TestAssetExecuteResizesPNGImage(t *testing.T) {
	asset := &Asset{
		Files:  []string{"images/cover.png"},
		Output: "cover-small.png",
		Filters: []Filter{
			&ImageFilter{Width: 40, Fit: imageFitInside, Quality: 85, Format: "png"},
		},
	}
	w := writer.NewMemoryWriter()

	err := asset.Execute(context.Background(), fstest.MapFS{
		"images/cover.png": &fstest.MapFile{Data: testPNG(t, 100, 80)},
	}, w)
	require.NoError(t, err)

	img := decodeOutputImage(t, w, "cover-small.png")
	assert.Equal(t, image.Rect(0, 0, 40, 32), img.Bounds())
}

func TestAssetExecuteReturnsErrorForMultipleImageMatches(t *testing.T) {
	asset := &Asset{
		Files:  []string{"images/*.jpg"},
		Output: "images/cover-small.jpg",
		Filters: []Filter{
			&ImageFilter{Fit: imageFitInside, Quality: 85, Format: "jpg"},
		},
	}

	err := asset.Execute(context.Background(), fstest.MapFS{
		"images/a.jpg": &fstest.MapFile{Data: testJPEG(t, 100, 80)},
		"images/b.jpg": &fstest.MapFile{Data: testJPEG(t, 100, 80)},
	}, writer.NewMemoryWriter())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image filter requires exactly one matched file")
}

func TestAssetExecuteReturnsErrorForImageFilterOnCSS(t *testing.T) {
	asset := &Asset{
		Files:  []string{"css/app.css"},
		Output: "images/app.jpg",
		Filters: []Filter{
			&ImageFilter{Fit: imageFitInside, Quality: 85, Format: "jpg"},
		},
	}

	err := asset.Execute(context.Background(), fstest.MapFS{
		"css/app.css": &fstest.MapFile{Data: []byte("body { color: red; }")},
	}, writer.NewMemoryWriter())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "image filter requires an image input")
}

func testJPEG(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, testImage(width, height), &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, testImage(width, height)))
	return buf.Bytes()
}

func testImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 120, A: 255})
		}
	}
	return img
}

func decodeOutputImage(t *testing.T, w *writer.MemoryWriter, file string) image.Image {
	t.Helper()

	if file[0] != '/' {
		file = "/" + file
	}
	f, err := w.Open(file)
	require.NoError(t, err)
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	require.NoError(t, err)

	img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	return img
}
