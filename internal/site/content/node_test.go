package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/honmaple/snow/internal/site/content/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNodePreservesRawContent(t *testing.T) {
	root := t.TempDir()
	contentDir := filepath.Join(root, "content")
	require.NoError(t, os.MkdirAll(filepath.Join(contentDir, "blog"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(contentDir, "hello.md"), []byte("# Hello"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(contentDir, "blog", "_index.md"), []byte("# Blog"), 0644))

	processor := newAssetTestProcessor(t, root, &parser.Result{
		Content:    "<p>Hello</p>",
		RawContent: "# Hello",
	})
	contentFS := os.DirFS(contentDir)

	page, err := processor.ParsePage(contentFS, "hello.md", false)
	require.NoError(t, err)
	assert.Equal(t, "# Hello", page.RawContent)

	section, err := processor.ParseSection(contentFS, "blog/_index.md")
	require.NoError(t, err)
	assert.Equal(t, "# Hello", section.RawContent)
}
