package assets

import (
	"fmt"
	"io/fs"
	"net/url"
	stdpath "path"
	"strings"

	"github.com/bep/godartsass/v2"
	"github.com/bep/golibsass/libsass"
)

const (
	sassCompilerLibSass  = "libscss"
	sassCompilerDartSass = "dartsass"
)

var allowedSassCompilers = map[string]bool{
	sassCompilerLibSass:  true,
	sassCompilerDartSass: true,
}

func normalizeSassCompiler(compiler string) (string, error) {
	compiler = strings.TrimSpace(compiler)
	if compiler == "" {
		return sassCompilerLibSass, nil
	}
	if !allowedSassCompilers[compiler] {
		return "", fmt.Errorf("unknown Sass compiler %q", compiler)
	}
	return compiler, nil
}

func sassSourceSyntax(file string) godartsass.SourceSyntax {
	switch stdpath.Ext(file) {
	case ".sass":
		return godartsass.SourceSyntaxSASS
	case ".css":
		return godartsass.SourceSyntaxCSS
	default:
		return godartsass.SourceSyntaxSCSS
	}
}

type dartSassImportResolver struct {
	assetsFS fs.FS
	dir      string
}

func (r *dartSassImportResolver) importCandidates(name string) []string {
	if stdpath.Ext(name) != "" {
		dir := stdpath.Dir(name)
		base := stdpath.Base(name)
		if dir == "." {
			dir = ""
		}
		return []string{
			name,
			stdpath.Join(dir, "_"+base),
		}
	}
	return []string{
		name + ".scss",
		name + ".sass",
		"_" + name + ".scss",
		"_" + name + ".sass",
	}
}

func (r *dartSassImportResolver) CanonicalizeURL(rawURL string) (string, error) {
	name := rawURL
	if u, err := url.Parse(rawURL); err == nil && u.Scheme == "file" {
		name = strings.TrimPrefix(u.Path, "/")
	}
	name = strings.TrimPrefix(name, "/")
	for _, candidate := range r.importCandidates(name) {
		path := stdpath.Join(r.dir, candidate)
		if _, err := fs.Stat(r.assetsFS, path); err == nil {
			return (&url.URL{
				Scheme: "snow-asset",
				Path:   "/" + path,
			}).String(), nil
		}
	}
	return "", nil
}

func (r *dartSassImportResolver) Load(canonicalizedURL string) (godartsass.Import, error) {
	u, err := url.Parse(canonicalizedURL)
	if err != nil {
		return godartsass.Import{}, err
	}
	if u.Scheme != "snow-asset" {
		return godartsass.Import{}, fmt.Errorf("%w: unsupported Dart Sass import URL %q", fs.ErrInvalid, canonicalizedURL)
	}

	path := strings.TrimPrefix(u.Path, "/")
	buf, err := fs.ReadFile(r.assetsFS, path)
	if err != nil {
		return godartsass.Import{}, err
	}
	return godartsass.Import{
		Content:      string(buf),
		SourceSyntax: sassSourceSyntax(path),
	}, nil
}

func (n *Asset) compileSass(assetsFS fs.FS, file string, buf []byte) ([]byte, error) {
	switch n.SassCompiler {
	case sassCompilerDartSass:
		return n.dartSass(assetsFS, file, buf)
	default:
		return n.libscss(assetsFS, file, buf)
	}
}

func (n *Asset) libscss(assetsFS fs.FS, file string, buf []byte) ([]byte, error) {
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

func (n *Asset) dartSass(assetsFS fs.FS, file string, buf []byte) ([]byte, error) {
	transpiler, err := godartsass.Start(godartsass.Options{})
	if err != nil {
		return nil, err
	}
	defer transpiler.Close()

	result, err := transpiler.Execute(godartsass.Args{
		Source:       string(buf),
		URL:          (&url.URL{Scheme: "file", Path: "/" + file}).String(),
		SourceSyntax: sassSourceSyntax(file),
		ImportResolver: &dartSassImportResolver{
			assetsFS: assetsFS,
			dir:      stdpath.Dir(file),
		},
	})
	if err != nil {
		return nil, err
	}
	return []byte(result.CSS), nil
}
