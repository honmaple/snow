package assets

import (
	"bytes"
	"context"
	"io/fs"
	stdpath "path"

	"github.com/honmaple/snow/internal/core"
)

type (
	Asset struct {
		Files        []string
		Filters      []Filter
		SassCompiler SassCompiler
		Output       string
		ShowVersion  bool
	}
	Filter interface {
		Name() string
		Execute([]byte) ([]byte, error)
	}
)

func (n *Asset) matchedFiles(assetsFS fs.FS) ([]string, error) {
	var results []string
	for _, file := range n.Files {
		matchedFiles, err := fs.Glob(assetsFS, file)
		if err != nil {
			return nil, err
		}
		results = append(results, matchedFiles...)
	}
	return results, nil
}

func (n *Asset) compileSass(assetsFS fs.FS, file string, buf []byte) ([]byte, error) {
	return n.SassCompiler.Compile(assetsFS, file, buf)
}

func (n *Asset) Execute(ctx context.Context, assetsFS fs.FS, writer core.Writer) error {
	if n.hasImageFilter() {
		buf, err := n.executeImage(assetsFS)
		if err != nil {
			return err
		}
		return writer.WriteFile(ctx, n.Output, bytes.NewReader(buf))
	}

	var (
		b bytes.Buffer
	)
	// 先合并再压缩
	matchedFiles, err := n.matchedFiles(assetsFS)
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

	buf := b.Bytes()
	for _, filter := range n.Filters {
		result, err := filter.Execute(buf)
		if err != nil {
			return err
		}
		buf = result
	}

	b.Reset()
	b.Write(buf)
	return writer.WriteFile(ctx, n.Output, &b)
}
