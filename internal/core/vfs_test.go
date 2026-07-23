package core

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func writeTestFile(t *testing.T, path string, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newTestContext(t *testing.T, root string, conf *Config) *Context {
	t.Helper()
	t.Chdir(root)
	ctx, err := NewContext(conf)
	if err != nil {
		t.Fatal(err)
	}
	return ctx
}

func TestVirtualFSMountsUseFixedNames(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "content", "index.md"), "home")
	writeTestFile(t, filepath.Join(root, "static", "style.css"), "body{}")
	writeTestFile(t, filepath.Join(root, "custom-content", "index.md"), "custom")
	writeTestFile(t, filepath.Join(root, "public", "style.css"), "custom{}")

	conf := DefaultConfig()
	conf.Set("content_dir", "custom-content")
	conf.Set("static_dir", "public")
	ctx := newTestContext(t, root, conf)

	rootFS := ctx.FS
	content, err := fs.ReadFile(rootFS, "content/index.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "home" {
		t.Fatalf("expected content mount to use fixed content dir, got %q", content)
	}

	static, err := fs.ReadFile(rootFS, "static/style.css")
	if err != nil {
		t.Fatal(err)
	}
	if string(static) != "body{}" {
		t.Fatalf("expected static mount to use fixed static dir, got %q", static)
	}

	content, err = fs.ReadFile(ctx.FS, "content/index.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "home" {
		t.Fatalf("expected ctx.FS to act as root FS, got %q", content)
	}
}

func TestVirtualFSRootUsesCurrentDirAndMounts(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "README.md"), "readme")
	writeTestFile(t, filepath.Join(root, "content", "_index.md"), "home")
	writeTestFile(t, filepath.Join(root, "external", "docs", "index.md"), "docs")
	writeTestFile(t, filepath.Join(root, "external", "manual", "index.md"), "manual")

	ctx := newTestContext(t, root, DefaultConfig())
	if err := ctx.FS.Mount(os.DirFS(filepath.Join(root, "external", "docs")), "content/docs/project-name"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.FS.Mount(os.DirFS(filepath.Join(root, "external", "manual")), "manual"); err != nil {
		t.Fatal(err)
	}

	got, err := fs.ReadFile(ctx.FS, "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "readme" {
		t.Fatalf("expected root fs to open current directory file, got %q", got)
	}

	entries, err := fs.ReadDir(ctx.FS, ".")
	if err != nil {
		t.Fatal(err)
	}
	names := dirEntryNames(entries)
	for _, name := range []string{"README.md", "content", "external", "manual"} {
		if !slices.Contains(names, name) {
			t.Fatalf("expected root entries to include %q, got %#v", name, names)
		}
	}
	if slices.Contains(names, MountStatic) {
		t.Fatalf("expected absent fixed mount %q not to be synthesized, got %#v", MountStatic, names)
	}

	entries, err = fs.ReadDir(ctx.FS, "content")
	if err != nil {
		t.Fatal(err)
	}
	names = dirEntryNames(entries)
	for _, name := range []string{"_index.md", "docs"} {
		if !slices.Contains(names, name) {
			t.Fatalf("expected content entries to include %q, got %#v", name, names)
		}
	}
}

func TestVirtualFSMountPriority(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "templates", "base.html"), "project")
	writeTestFile(t, filepath.Join(root, "themes", "demo", "templates", "base.html"), "theme")
	writeTestFile(t, filepath.Join(root, "themes", "demo", "data123", "site.yaml"), "name: data123")
	writeTestFile(t, filepath.Join(root, "themes", "demo", "data", "site.yaml"), "name: theme")
	writeTestFile(t, filepath.Join(root, "themes", "demo", "theme.yaml"), "name: demo")

	conf := DefaultConfig()
	conf.Set("theme", "demo")
	ctx := newTestContext(t, root, conf)

	templatesFS, err := ctx.GetFS(MountTemplates, true, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fs.ReadFile(templatesFS, "base.html")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "project" {
		t.Fatalf("expected project template to override theme template, got %q", got)
	}

	dataFS, err := ctx.GetFS("data", true, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err = fs.ReadFile(dataFS, "site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "name: theme" {
		t.Fatalf("expected data mount to fall back to theme data, got %q", got)
	}

	data123FS, err := ctx.GetFS("data123", true, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err = fs.ReadFile(data123FS, "site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "name: data123" {
		t.Fatalf("expected GetFS to fall back to theme data123, got %q", got)
	}

	got, err = fs.ReadFile(ctx.FS, "themes/demo/theme.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "name: demo" {
		t.Fatalf("expected root fs to expose current directory theme file, got %q", got)
	}
}

func dirEntryNames(entries []fs.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestGetFSKeepsInternalTemplatesCompatibility(t *testing.T) {
	root := t.TempDir()
	ctx := newTestContext(t, root, DefaultConfig())

	mountFS, err := ctx.GetFS(MountTemplates, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(mountFS, "_partials/base.html"); err != nil {
		t.Fatalf("expected templates mount to include embedded templates: %v", err)
	}

	templatesFS, err := ctx.GetFS(MountTemplates, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(templatesFS, "_partials/base.html"); err != nil {
		t.Fatalf("expected GetFS templates with internal=true to include embedded templates: %v", err)
	}
	if _, err := fs.ReadFile(templatesFS, "internal/_partials/base.html"); err != nil {
		t.Fatalf("expected GetFS templates with internal=true to expose embedded templates under internal prefix: %v", err)
	}

	staticFS, err := ctx.GetFS(MountStatic, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(staticFS, "internal/css/style.css"); err == nil {
		t.Fatal("expected internal theme not to expose embedded static files")
	}
}

func TestGetFSWithoutInternalDoesNotExposeInternalTheme(t *testing.T) {
	root := t.TempDir()
	ctx := newTestContext(t, root, DefaultConfig())

	templatesFS, err := ctx.GetFS(MountTemplates, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(templatesFS, "_partials/base.html"); err == nil {
		t.Fatal("expected templates with internal=false not to fall back to embedded templates")
	}
	if _, err := fs.ReadFile(templatesFS, "internal/_partials/base.html"); err == nil {
		t.Fatal("expected templates with internal=false not to expose embedded templates under internal prefix")
	}

	staticFS, err := ctx.GetFS(MountStatic, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(staticFS, "css/style.css"); err == nil {
		t.Fatal("expected static with internal=false not to fall back to embedded static")
	}
	if _, err := fs.ReadFile(staticFS, "internal/css/style.css"); err == nil {
		t.Fatal("expected static with internal=false not to expose embedded static under internal prefix")
	}
}

func TestGetFSStaticDoesNotFallBackToInternalTheme(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "themes", "demo", "theme.yaml"), "name: demo")

	conf := DefaultConfig()
	conf.Set("theme", "demo")
	ctx := newTestContext(t, root, conf)

	staticFS, err := ctx.GetFS(MountStatic, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ReadFile(staticFS, "css/style.css"); err == nil {
		t.Fatal("expected static mount not to fall back to embedded static")
	}
	if _, err := fs.ReadFile(staticFS, "internal/css/style.css"); err == nil {
		t.Fatal("expected static mount not to expose embedded static under internal prefix")
	}
}

func TestVirtualFSEmptyMount(t *testing.T) {
	root := t.TempDir()
	ctx := newTestContext(t, root, DefaultConfig())

	assetsFS, err := ctx.GetFS("assets", true, false)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(assetsFS, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty assets mount, got %d entries", len(entries))
	}

}

func TestVirtualFSMountFS(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "content", "_index.md"), "home")
	writeTestFile(t, filepath.Join(root, "external", "docs", "index.md"), "docs")
	writeTestFile(t, filepath.Join(root, "external", "data123", "site.yaml"), "data123")

	ctx := newTestContext(t, root, DefaultConfig())
	if err := ctx.FS.Mount(os.DirFS(filepath.Join(root, "external", "docs")), "content/docs/project-name"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.FS.Mount(testSingleFileFS{name: "style.css", data: []byte("style")}, "static/style.css"); err != nil {
		t.Fatal(err)
	}
	if err := ctx.FS.Mount(os.DirFS(filepath.Join(root, "external", "data123")), "data123"); err != nil {
		t.Fatal(err)
	}

	contentFS, err := ctx.GetFS(MountContent, false, false)
	if err != nil {
		t.Fatal(err)
	}
	got, err := fs.ReadFile(contentFS, "docs/project-name/index.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "docs" {
		t.Fatalf("expected mounted content file, got %q", got)
	}

	staticFS, err := ctx.GetFS(MountStatic, true, true)
	if err != nil {
		t.Fatal(err)
	}
	got, err = fs.ReadFile(staticFS, "style.css")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "style" {
		t.Fatalf("expected mounted static file, got %q", got)
	}

	got, err = fs.ReadFile(ctx.FS, "content/docs/project-name/index.md")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "docs" {
		t.Fatalf("expected mounted file through root fs, got %q", got)
	}

	got, err = fs.ReadFile(ctx.FS, "data123/site.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "data123" {
		t.Fatalf("expected root fs to open mounted non-fixed directory, got %q", got)
	}

	entries, err := fs.ReadDir(contentFS, "docs")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "project-name" || !entries[0].IsDir() {
		t.Fatalf("expected mounted docs/project-name directory, got %#v", entries)
	}
}

func TestVirtualFSMountStrategies(t *testing.T) {
	tests := []struct {
		name         string
		strategy     string
		sourceIndex  string
		wantIndex    string
		wantIndexErr bool
	}{
		{
			name:        "mount first",
			strategy:    MountStrategyMount,
			sourceIndex: "source index",
			wantIndex:   "source index",
		},
		{
			name:        "base first",
			strategy:    MountStrategyBase,
			sourceIndex: "source index",
			wantIndex:   "base index",
		},
		{
			name:         "override",
			strategy:     MountStrategyOverride,
			wantIndexErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeTestFile(t, filepath.Join(root, "content", "docs", "_index.md"), "base index")
			writeTestFile(t, filepath.Join(root, "content", "docs", "base.md"), "base page")
			writeTestFile(t, filepath.Join(root, "content", "other.md"), "other page")
			writeTestFile(t, filepath.Join(root, "external", "docs", "mount.md"), "mount page")
			if tt.sourceIndex != "" {
				writeTestFile(t, filepath.Join(root, "external", "docs", "_index.md"), tt.sourceIndex)
			}

			ctx := newTestContext(t, root, DefaultConfig())
			err := ctx.FS.MountWithStrategy(os.DirFS(filepath.Join(root, "external", "docs")), "content/docs", tt.strategy)
			if err != nil {
				t.Fatal(err)
			}

			got, err := fs.ReadFile(ctx.FS, "content/docs/_index.md")
			if tt.wantIndexErr {
				if err == nil {
					t.Fatalf("expected _index.md to be hidden by override mount, got %q", got)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != tt.wantIndex {
					t.Fatalf("expected _index.md %q, got %q", tt.wantIndex, got)
				}
			}

			got, err = fs.ReadFile(ctx.FS, "content/docs/mount.md")
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != "mount page" {
				t.Fatalf("expected mount file, got %q", got)
			}

			got, err = fs.ReadFile(ctx.FS, "content/other.md")
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != "other page" {
				t.Fatalf("expected sibling outside override target to stay visible, got %q", got)
			}
		})
	}
}

func TestVirtualFSMountRejectsInvalidStrategy(t *testing.T) {
	root := t.TempDir()
	ctx := newTestContext(t, root, DefaultConfig())
	err := ctx.FS.MountWithStrategy(os.DirFS(root), "content/docs", "unknown")
	if err == nil {
		t.Fatal("expected unknown mount strategy to be rejected")
	}
}

func TestVirtualFSReplaceFileMountOnlyCoversExactTarget(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "static", "style.css", "nested.css"), "base nested")

	ctx := newTestContext(t, root, DefaultConfig())
	err := ctx.FS.MountWithStrategy(testSingleFileFS{name: "style.css", data: []byte("source style")}, "static/style.css", MountStrategyOverride)
	if err != nil {
		t.Fatal(err)
	}

	got, err := fs.ReadFile(ctx.FS, "static/style.css")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "source style" {
		t.Fatalf("expected exact file target to use mounted file, got %q", got)
	}

	got, err = fs.ReadFile(ctx.FS, "static/style.css/nested.css")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "base nested" {
		t.Fatalf("expected file mount not to override nested path, got %q", got)
	}
}

func TestVirtualFSMountRejectsInvalidTarget(t *testing.T) {
	root := t.TempDir()
	ctx := newTestContext(t, root, DefaultConfig())
	source := os.DirFS(root)

	for _, target := range []string{
		"",
		".",
		"/content/docs",
		"./content/docs",
		"content//docs",
		"content/docs/",
		"../content/docs",
		"content/../docs",
	} {
		if err := ctx.FS.Mount(source, target); err == nil {
			t.Fatalf("expected mount target %q to be rejected", target)
		}
	}
}

type testSingleFileFS struct {
	name string
	data []byte
}

func (f testSingleFileFS) Open(name string) (fs.File, error) {
	if name != "." {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &testSingleFile{
		Reader: bytes.NewReader(f.data),
		name:   f.name,
		size:   int64(len(f.data)),
	}, nil
}

type testSingleFile struct {
	*bytes.Reader
	name string
	size int64
}

func (f *testSingleFile) Stat() (fs.FileInfo, error) {
	return testSingleFileInfo{name: f.name, size: f.size}, nil
}

func (f *testSingleFile) Close() error {
	return nil
}

type testSingleFileInfo struct {
	name string
	size int64
}

func (i testSingleFileInfo) Name() string       { return i.name }
func (i testSingleFileInfo) Size() int64        { return i.size }
func (i testSingleFileInfo) Mode() fs.FileMode  { return 0o444 }
func (i testSingleFileInfo) ModTime() time.Time { return time.Time{} }
func (i testSingleFileInfo) IsDir() bool        { return false }
func (i testSingleFileInfo) Sys() any           { return nil }
