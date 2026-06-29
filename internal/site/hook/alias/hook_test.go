package alias

import (
	"context"
	"io"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/writer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAliases(t *testing.T) {
	aliases, err := parseAliases([]string{
		"/old-url/:/new-url/",
		"/external/:https://example.com/new/",
		"",
	})

	require.NoError(t, err)
	require.Len(t, aliases, 2)
	assert.Equal(t, Alias{From: "/old-url/", To: "/new-url/"}, aliases[0])
	assert.Equal(t, Alias{From: "/external/", To: "https://example.com/new/"}, aliases[1])
}

func TestParseAliasesRequiresSeparator(t *testing.T) {
	_, err := parseAliases([]string{"/old-url/"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected old:new")
}

func TestParseAliasesRequiresCleanOldURL(t *testing.T) {
	tests := []string{
		"../old/:/new/",
		"./old/:/new/",
		"/old/../new/:/new/",
		"//old/:/new/",
		"/old/?from=1:/new/",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			_, err := parseAliases([]string{tt})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid alias")
		})
	}
}

func TestAliasHookWritesRedirectFiles(t *testing.T) {
	conf := core.DefaultConfig()
	conf.Set("base_url", "https://example.com")
	conf.Set("hooks.alias.option", []string{
		"/old-url/:/new-url/",
		"/old.html:/new.html",
	})
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	hook, err := New(ctx)
	require.NoError(t, err)

	w := writer.NewMemoryWriter()
	require.NoError(t, hook.AfterBuild(context.Background(), w))

	assertRedirectFile(t, w, "/old-url/index.html", `https://example.com/new-url/`)
	assertRedirectFile(t, w, "/old.html", `https://example.com/new.html`)
}

func assertRedirectFile(t *testing.T, w *writer.MemoryWriter, file string, target string) {
	t.Helper()

	f, err := w.Open(file)
	require.NoError(t, err)
	defer f.Close()

	b, err := io.ReadAll(f)
	require.NoError(t, err)
	assert.Contains(t, string(b), `content="0; url=`+target+`"`)
	assert.Contains(t, string(b), `rel="canonical" href="`+target+`"`)
}
