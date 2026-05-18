package assets

import (
	"bytes"
	"context"
	"io/fs"
	stdpath "path"

	"github.com/bep/golibsass/libsass"
	"github.com/honmaple/snow/internal/core"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type Asset struct {
	Files       []string
	Filters     []string
	Output      string
	ShowVersion bool
}

func getFirstKey[K comparable, V any](m map[K]V) (key K, ok bool) {
	for k := range m {
		return k, true
	}
	return key, false
}

func (n *Asset) filter(name string, buf []byte) (result []byte, err error) {
	switch name {
	// case "libscss":
	//	result, err = n.libscss(assetsFS, file, buf, opt)
	case "cssmin":
		result, err = n.cssmin(buf)
	case "jsmin":
		result, err = n.jsmin(buf)
	}
	return result, err
}

func (n *Asset) cssmin(buf []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("css", css.Minify)

	return m.Bytes("css", buf)
}

func (n *Asset) jsmin(buf []byte) ([]byte, error) {
	m := minify.New()
	m.AddFunc("js", js.Minify)

	return m.Bytes("js", buf)
}

func (n *Asset) libscss(assetsFS fs.FS, file string, buf []byte, opt map[string]any) ([]byte, error) {
	dir := stdpath.Dir(file)

	opts := libsass.Options{}
	opts.ImportResolver = func(url string, prev string) (newURL string, body string, resolved bool) {
		if stdpath.Ext(url) == "" {
			urls := []string{
				url + ".scss",
				url + ".sass",
				"_" + url + ".scss",
				"_" + url + ".sass",
			}
			for _, u := range urls {
				if _, err := fs.Stat(assetsFS, stdpath.Join(dir, u)); err == nil {
					url = u
					break
				}
			}
		}
		b, err := fs.ReadFile(assetsFS, stdpath.Join(dir, url))
		if err != nil {
			return url, "", false
		}
		return url, string(b), true
	}

	transpiler, err := libsass.New(opts)
	if err != nil {
		return nil, err
	}

	result, err := transpiler.Execute(string(buf))
	if err != nil {
		return nil, err
	}
	return []byte(result.CSS), nil
}

func (n *Asset) Execute(ctx context.Context, assetsFS fs.FS, writer core.Writer) error {
	var (
		b bytes.Buffer
	)
	// 先合并再压缩
	for _, file := range n.Files {
		matchedFiles, err := fs.Glob(assetsFS, file)
		if err != nil {
			return err
		}

		for _, matchedFile := range matchedFiles {
			buf, err := fs.ReadFile(assetsFS, matchedFile)
			if err != nil {
				return err
			}
			switch stdpath.Ext(matchedFile) {
			case ".scss", ".sass":
				result, err := n.libscss(assetsFS, matchedFile, buf, nil)
				if err != nil {
					return err
				}
				b.Write(result)
				b.WriteString("\n")
			case ".css":
				b.Write(buf)
				b.WriteString("\n")
			case ".js":
				b.Write(buf)
				b.WriteString("\n;\n")
			}
		}
	}

	buf := b.Bytes()
	for _, name := range n.Filters {
		result, err := n.filter(name, buf)
		if err != nil {
			return err
		}
		buf = result
	}

	b.Reset()
	b.Write(buf)
	return writer.WriteFile(ctx, n.Output, &b)
}
