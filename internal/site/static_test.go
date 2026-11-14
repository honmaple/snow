package site

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/writer"
	"github.com/stretchr/testify/require"
)

func TestResolveStaticPath(t *testing.T) {
	site := &Site{}
	paths := map[string]string{
		"css":                "static/css",
		"css/vendor":         "vendor",
		"css/vendor/app.css": "app.css",
		"css/kit":            "kit",
		"readme.md":          "docs/README.md",
	}
	for _, tt := range []struct {
		path string
		want string
	}{
		{"CSS/components/Deep/App.css", "static/css/components/Deep/App.css"},
		{"CSS/Vendor/Deep/App.css", "vendor/Deep/App.css"},
		{"CSS/Vendor/App.css", "app.css"},
		{"css/\u212ait/Deep/App.css", "kit/Deep/App.css"},
		{"README.md", "docs/README.md"},
		{"README/notes.txt", "README/notes.txt"},
		{"css-extra/App.css", "css-extra/App.css"},
		{"robots.txt", "robots.txt"},
	} {
		require.Equal(t, tt.want, site.resolveStaticPath(tt.path, paths), tt.path)
	}
}

func TestBuildStaticPaths(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	config := `theme: test
statics:
  css: static/css
  css/vendor: vendor
  favicon.ico: icons/site.ico
  README.md: docs/README.md
ignored_static: [css/ignored.css]
`
	require.NoError(t, os.WriteFile("config.yaml", []byte(config), 0644))
	require.NoError(t, os.MkdirAll("themes/test/static/css", 0755))
	require.NoError(t, os.WriteFile("themes/test/static/css/theme.css", []byte("theme"), 0644))
	require.NoError(t, os.WriteFile("themes/test/static/css/app.css", []byte("theme app"), 0644))

	conf := core.NewConfig()
	require.NoError(t, conf.LoadFromFile("config.yaml"))
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)
	require.NoError(t, ctx.FS.Mount(fstest.MapFS{
		"css/app.css":          {Data: []byte("app")},
		"css/vendor/reset.css": {Data: []byte("reset")},
		"css/ignored.css":      {Data: []byte("ignored")},
		"css-extra/extra.css":  {Data: []byte("extra")},
		"favicon.ico":          {Data: []byte("icon")},
		"README.md":            {Data: []byte("readme")},
		"README/notes.txt":     {Data: []byte("notes")},
		"robots.txt":           {Data: []byte("robots")},
	}, core.MountStatic))

	site := &Site{ctx: ctx}
	output := t.TempDir()
	require.NoError(t, site.BuildStatic(t.Context(), writer.NewDiskWriter(output)))
	for path, want := range map[string]string{
		"static/css/app.css":   "app",
		"static/css/theme.css": "theme",
		"vendor/reset.css":     "reset",
		"css-extra/extra.css":  "extra",
		"icons/site.ico":       "icon",
		"docs/README.md":       "readme",
		"README/notes.txt":     "notes",
		"robots.txt":           "robots",
	} {
		got, err := os.ReadFile(filepath.Join(output, path))
		require.NoError(t, err)
		require.Equal(t, want, string(got), path)
	}
	for _, path := range []string{"css/app.css", "favicon.ico", "static/css/ignored.css"} {
		_, err := os.Stat(filepath.Join(output, path))
		require.ErrorIs(t, err, fs.ErrNotExist)
	}
}

func TestBuildStaticSharedOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, tt := range []struct {
		target string
		want   string
		err    string
	}{
		{"robots.txt", "robots", ""},
		{"robots.txt/app.css", "app", "robots.txt"},
	} {
		conf := core.NewConfig()
		conf.Set("statics.app.css", tt.target)
		ctx, err := core.NewContext(conf)
		require.NoError(t, err)
		require.NoError(t, ctx.FS.Mount(fstest.MapFS{
			"app.css":    {Data: []byte("app")},
			"robots.txt": {Data: []byte("robots")},
		}, core.MountStatic))

		site := &Site{ctx: ctx}
		output := t.TempDir()
		err = site.BuildStatic(t.Context(), writer.NewDiskWriter(output))
		if tt.err == "" {
			require.NoError(t, err)
		} else {
			require.ErrorContains(t, err, tt.err)
		}
		got, err := os.ReadFile(filepath.Join(output, tt.target))
		require.NoError(t, err)
		require.Equal(t, tt.want, string(got))
	}
}
