package loader

import (
	"fmt"
	"io/fs"
	"os"
	stdpath "path"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/content/types"
)

func (d *DiskLoader) isIgnoredContent(path string, isDir bool) bool {
	// 忽略以.或者_开头的文件或目录，不要忽略_index.md
	if basename := filepath.Base(path); !strings.HasPrefix(basename, "_index.") && (strings.HasPrefix(basename, "_") || strings.HasPrefix(basename, ".")) {
		return true
	}

	matchPath := strings.TrimPrefix(path, d.ctx.GetContentDir()+"/")
	if isDir {
		matchPath = matchPath + "/"
	}
	for _, pattern := range d.ctx.Config.GetStringSlice("ignored_content") {
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

func (d *DiskLoader) findIndexFiles(fullpath string, prefix string) []string {
	// 如果有多个扩展: index.md, index.org只返回第一个
	allowedFiles := make(map[string]bool)
	for _, ext := range d.parser.SupportedExtensions() {
		allowedFiles[prefix+ext] = true
		for lang := range d.ctx.Locales {
			allowedFiles[prefix+lang+ext] = true
		}
	}

	files, err := os.ReadDir(fullpath)
	if err != nil {
		return nil
	}

	results := make([]string, 0)
	resultMap := make(map[string]bool)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		nameWithoutExt := strings.TrimSuffix(name, filepath.Ext(name))
		if allowedFiles[name] && !resultMap[nameWithoutExt] {
			results = append(results, name)
			resultMap[nameWithoutExt] = true
		}
	}
	return results
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
	contentDir := d.ctx.GetContentDir()
	if contentDir == "" {
		return nil, fmt.Errorf("The content dir is null")
	}

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == contentDir {
			sectionFiles := d.findIndexFiles(path, "_index")
			if len(sectionFiles) > 0 {
				for _, file := range sectionFiles {
					d.insertSection(filepath.Join(path, file), true)
				}
				return nil
			}
			return nil
		}
		// 忽略指定的文件
		if d.isIgnoredContent(path, info.IsDir()) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			indexFiles := d.findIndexFiles(path, "index")
			if len(indexFiles) > 0 {
				for _, file := range indexFiles {
					d.insertPage(filepath.Join(path, file), true)
				}
				return fs.SkipDir
			}
			sectionFiles := d.findIndexFiles(path, "_index")
			if len(sectionFiles) > 0 {
				for _, file := range sectionFiles {
					d.insertSection(filepath.Join(path, file), false)
				}
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

	if err := filepath.WalkDir(contentDir, walkDir); err != nil {
		return nil, err
	}
	return d, nil
}
