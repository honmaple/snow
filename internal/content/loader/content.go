package loader

import (
	"fmt"
	"io/fs"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/content/types"
)

func (d *DiskLoader) isIgnoredContent(root, path string, info fs.DirEntry) bool {
	// 忽略以.或者_开头的文件或目录，不要忽略_index.md
	if basename := filepath.Base(path); !strings.HasPrefix(basename, "_index.") && (strings.HasPrefix(basename, "_") || strings.HasPrefix(basename, ".")) {
		return true
	}

	for _, pattern := range d.ctx.Config.GetStringSlice("ignored_content") {
		matchPath := strings.TrimPrefix(path, root+"/")
		if info.IsDir() {
			matchPath = matchPath + "/"
		}
		matched, err := filepath.Match(pattern, matchPath)
		if err != nil {
			d.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (d *DiskLoader) findFiles(path string, pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(path, pattern))
	if len(matches) == 0 {
		return nil
	}

	files := make([]string, 0)
	for _, m := range matches {
		if d.parserExts[filepath.Ext(m)] {
			files = append(files, m)
		}
	}
	return files
}

func (d *DiskLoader) loadFile(fullpath string) (*types.File, error) {
	relPath, err := filepath.Rel(d.ctx.GetContentDir(), fullpath)
	if err != nil {
		return nil, err
	}
	relPath = filepath.ToSlash(relPath)

	ext := stdpath.Ext(relPath)
	nameWithExt := stdpath.Base(relPath)
	nameWithoutExt := strings.TrimSuffix(nameWithExt, ext)

	dir := stdpath.Dir(relPath)
	if dir == "." {
		dir = ""
	}
	return &types.File{
		Path:     relPath,
		Dir:      dir,
		Ext:      ext,
		Name:     nameWithExt,
		BaseName: nameWithoutExt,
	}, nil
}

func (d *DiskLoader) Load() (types.Store, error) {
	root := d.ctx.GetContentDir()
	if root == "" {
		return nil, fmt.Errorf("The content dir is null")
	}

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			d.insertIndexSection(path)
			return nil
		}
		// 忽略指定的文件
		if d.isIgnoredContent(root, path, info) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			// index.md、index.en.md
			indexFiles := d.findFiles(path, "index.*")
			if len(indexFiles) > 0 {
				d.insertPage(filepath.Join(path, indexFiles[0]), true)
				return fs.SkipDir
			}
			sectionFiles := d.findFiles(path, "_index.*")
			if len(sectionFiles) > 0 {
				d.insertSection(filepath.Join(path, sectionFiles[0]))
				return nil
			}
			return nil
		}

		if d.parserExts[filepath.Ext(path)] {
			d.insertPage(path, false)
			return nil
		}
		return d.insertSectionAsset(path)
	}

	if err := filepath.WalkDir(root, walkDir); err != nil {
		return nil, err
	}
	return d, nil
}
