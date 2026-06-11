package assets

import (
	"bytes"
	"context"
	"io/fs"
	stdpath "path"

	"github.com/honmaple/snow/internal/core"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

type Asset struct {
	Files        []string
	Filters      []string
	SassCompiler string
	Output       string
	ShowVersion  bool
}

func (n *Asset) filter(name string, buf []byte) (result []byte, err error) {
	switch name {
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
				result, err := n.compileSass(assetsFS, matchedFile, buf)
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
