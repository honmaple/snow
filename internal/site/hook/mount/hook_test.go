package mount

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/stretchr/testify/require"
)

func TestNewMountsConfiguredLocalPaths(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	require.NoError(t, os.MkdirAll(filepath.Join(root, "external", "docs"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "external", "docs", "index.md"), []byte("docs"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "external", "style.css"), []byte("style"), 0o644))

	conf := core.DefaultConfig()
	conf.Set("hooks.mount.option", []map[string]any{
		{
			"source": filepath.Join(root, "external", "docs"),
			"target": "content/docs/project-name",
		},
		{
			"source": filepath.Join(root, "external", "style.css"),
			"target": "static/style.css",
		},
	})
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	_, err = New(ctx)
	require.NoError(t, err)

	contentFS, err := ctx.GetFS(core.MountContent, false, false)
	require.NoError(t, err)
	got, err := fs.ReadFile(contentFS, "docs/project-name/index.md")
	require.NoError(t, err)
	require.Equal(t, "docs", string(got))

	staticFS, err := ctx.GetFS(core.MountStatic, true, true)
	require.NoError(t, err)
	got, err = fs.ReadFile(staticFS, "style.css")
	require.NoError(t, err)
	require.Equal(t, "style", string(got))
}
