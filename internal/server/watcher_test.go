package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site"
	"github.com/honmaple/snow/internal/writer"
	"github.com/stretchr/testify/require"
)

func TestReloadStaticRebuildsSite(t *testing.T) {
	t.Chdir(t.TempDir())
	require.NoError(t, os.MkdirAll("static/css", 0755))
	const file = "static/css/app.css"
	require.NoError(t, os.WriteFile(file, []byte("before"), 0644))
	require.NoError(t, os.WriteFile("static/robots.txt", []byte("robots"), 0644))

	conf := core.NewConfig()
	conf.Set("language", "en")
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)
	build, err := site.New(ctx)
	require.NoError(t, err)
	s := &Server{site: build, fs: writer.NewMemoryWriter()}
	require.NoError(t, build.Build(t.Context(), s.fs))

	conf.Set("statics", map[string]any{"css": "static/css"})
	require.NoError(t, os.WriteFile(file, []byte("after"), 0644))
	info, err := os.Stat(file)
	require.NoError(t, err)
	require.NoError(t, s.reloadStatic("static", file, info))

	for path, want := range map[string]string{
		"/static/css/app.css": "after",
		"/robots.txt":         "robots",
	} {
		res := httptest.NewRecorder()
		s.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, http.StatusOK, res.Code, path)
		require.Equal(t, want, res.Body.String(), path)
	}
	for path, want := range map[string]int{"/css/app.css": http.StatusNotFound, "/": http.StatusOK} {
		res := httptest.NewRecorder()
		s.ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		require.Equal(t, want, res.Code, path)
	}
}
