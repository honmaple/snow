package loader

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/honmaple/snow/internal/content/types"
	"strings"
)

func (d *DiskLoader) findLang(path string, lang string) string {
	if lang == "" {
		filename := filepath.Base(path)
		// 1. 去掉最末尾的后缀 (例如 .md)
		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

		// 2. 再次获取当前的后缀 (即潜在的语言代码)
		langExt := filepath.Ext(nameWithoutExt) // 得到 .en 或 .fr
		if langExt != "" {
			// 3. 去掉点，得到 en 或 fr
			lang = strings.TrimPrefix(langExt, ".")
		}
	}
	if lang == "" || lang == d.ctx.GetDefaultLanguage() {
		return d.ctx.GetDefaultLanguage()
	}
	if _, ok := d.ctx.Locales[lang]; ok {
		return lang
	}
	return d.ctx.GetDefaultLanguage()
}

func (d *DiskLoader) findFiles(path string, pattern string) []string {
	matches, _ := filepath.Glob(filepath.Join(path, pattern))
	if len(matches) == 0 {
		return nil
	}

	files := make([]string, 0)
	for _, m := range matches {
		if d.parser.IsSupported(filepath.Ext(m)) {
			files = append(files, m)
		}
	}
	return files
}

func (d *DiskLoader) findSection(path string) *types.Section {
	if path == "." {
		result, _ := d.sections.Find("@home")
		return result
	}

	section, ok := d.sections.Find(path)
	if ok {
		return section
	}
	return d.findSection(filepath.Dir(path))
}

func (d *DiskLoader) loadContents() (types.Store, error) {
	root := d.ctx.GetContentDir()
	if root == "" {
		return nil, fmt.Errorf("The content dir is null")
	}

	walkDir := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			d.insertSection(path, true)
			return nil
		}
		// 忽略以.或者_开头的文件或目录
		if basename := filepath.Base(path); strings.HasPrefix(basename, "_") || strings.HasPrefix(basename, ".") {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		// 忽略指定的文件
		for _, pattern := range d.ctx.Config.GetStringSlice("ignored_content") {
			matchPath := strings.TrimPrefix(path, root+"/")
			if info.IsDir() {
				matchPath = matchPath + "/"
			}
			matched, err := filepath.Match(pattern, matchPath)
			if err != nil {
				d.ctx.Logger.Warnf("The pattern %s match %s err: %s", pattern, path, err)
			}
			if matched {
				if info.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		if info.IsDir() {
			indexFiles := d.findFiles(path, "index.*")
			if len(indexFiles) > 0 {
				d.insertPage(filepath.Join(path, indexFiles[0]), true)
				return fs.SkipDir
			}
			sectionFiles := d.findFiles(path, "_index.*")
			if len(sectionFiles) > 0 {
				d.insertSection(filepath.Join(path, sectionFiles[0]), false)
				return nil
			}
			return nil
		}

		if d.parser.IsSupported(filepath.Ext(path)) {
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
