package loader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/honmaple/snow/internal/static/types"
	"github.com/honmaple/snow/internal/utils"
)

func (d *DiskLoader) Statics() types.Statics {
	return d.statics.List()
}

func (d *DiskLoader) GetStatic(path string) *types.Static {
	result, _ := d.statics.Find(path)
	return result
}

func (d *DiskLoader) GetStaticURL(path string) string {
	result := d.GetStatic(path)
	if result == nil {
		return ""
	}
	return result.Permalink
}

func (d *DiskLoader) getStaticPath(file string, isTheme bool) string {
	prefix := "statics."
	if isTheme {
		prefix = "statics.@theme/"
	}

	strs := strings.Split(file, "/")
	for i := len(strs); i > 0; i-- {
		path := strings.Join(strs[:i], "/")

		customPath := d.ctx.Config.GetString(fmt.Sprintf("%s%s.path", prefix, path))
		if customPath != "" {
			if path == file {
				return customPath
			}
			if strings.HasSuffix(customPath, "/") {
				return customPath + filepath.Base(path) + strings.TrimPrefix(file, path)
			}
			return customPath + strings.TrimPrefix(file, path)
		}
	}
	return ""
}

func (d *DiskLoader) insertStatic(file string) error {
	static := &types.Static{
		File: file,
	}

	customPath := d.getStaticPath(file, false)
	if customPath == "" {
		customPath = "static/{path}/{filename}"
	}

	outputPath := utils.StringReplace(customPath, map[string]string{
		"{path}":     filepath.Dir(file),
		"{filename}": filepath.Base(file),
	})

	if strings.HasSuffix(outputPath, "/") {
		static.Path = filepath.Join(outputPath, filepath.Base(file))
	} else {
		static.Path = outputPath
	}
	static.Path = d.ctx.GetRelURL(static.Path)
	static.Permalink = d.ctx.GetURL(static.Path)

	if d.hook != nil {
		static = d.hook.HandleStatic(static)
		if static == nil {
			return nil
		}
	}

	d.statics.Add(file, static)
	return nil
}

func (d *DiskLoader) insertThemeStatic(file string) error {
	// static/css/custom.css -> @theme/static/css/custom.css
	static := &types.Static{
		File: "@theme/" + file,
	}
	customPath := d.getStaticPath(file, true)
	if customPath == "" {
		customPath = "static/{path}/{filename}"
	}

	outputPath := utils.StringReplace(customPath, map[string]string{
		"{path}":     filepath.Dir(file),
		"{filename}": filepath.Base(file),
	})

	if strings.HasSuffix(outputPath, "/") {
		static.Path = filepath.Join(outputPath, filepath.Base(file))
	} else {
		static.Path = outputPath
	}
	static.Path = d.ctx.GetRelURL(static.Path)
	static.Permalink = d.ctx.GetURL(static.Path)

	if d.hook != nil {
		static = d.hook.HandleStatic(static)
		if static == nil {
			return nil
		}
	}

	d.statics.Add(static.File, static)
	return nil
}
