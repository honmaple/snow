/*********************************************************************************
 Copyright © 2019 jianglin
 File Name: template.go
 Author: jianglin
 Email: mail@honmaple.com
 Created: 2019-08-29 14:49:27 (CST)
 Last Update: Saturday 2019-09-07 02:39:05 (CST)
		  By:
 Description:
 *********************************************************************************/
package core

import (
	"github.com/flosch/pongo2"
	"path/filepath"
)

type Template struct {
	Name       string `toml:"name"`
	Pagination int64  `toml:"pagination"`
	SavePath   string `toml:"save_path"`

	Page       *Template `toml:"page"`
	Article    *Template `toml:"article"`
	Index      *Template `toml:"index"`
	Tag        *Template `toml:"tag"`
	Author     *Template `toml:"author"`
	Category   *Template `toml:"category"`
	Tags       *Template `toml:"tags"`
	Authors    *Template `toml:"authors"`
	Categories *Template `toml:"categories"`
	Archives   *Template `toml:"archives"`
}

var template = &Template{
	Pagination: 10,
	Index: &Template{
		Name:       "index.html",
		Pagination: 0,
		SavePath:   "index{number}.html",
	},
	Tags: &Template{
		Name:       "tags.html",
		Pagination: 0,
		SavePath:   "tags.html",
	},
	Categories: &Template{
		Name:       "categories.html",
		Pagination: 0,
		SavePath:   "categories.html",
	},
	Authors: &Template{
		Name:       "authors.html",
		Pagination: 0,
		SavePath:   "authors.html",
	},
	Archives: &Template{
		Name:       "archives.html",
		Pagination: 0,
		SavePath:   "archives.html",
	},
	Tag: &Template{
		Name:       "tag.html",
		Pagination: 0,
		SavePath:   "tag/{slug}/index{number}.html",
	},
	Category: &Template{
		Name:       "category.html",
		Pagination: 0,
		SavePath:   "category/{slug}/index{number}.html",
	},
	Author: &Template{
		Name:       "author.html",
		Pagination: 0,
		SavePath:   "author/{slug}/index{number}.html",
	},
	Article: &Template{
		Name:       "article.html",
		Pagination: 0,
		SavePath:   "articles/{date:2006}/{date:01}/{slug}.html",
	},
	Page: &Template{
		Name:       "page.html",
		Pagination: 0,
		SavePath:   "pages/{slug}.html",
	},
}

var templateRedirect = map[string]string{
	"tag.html":      "index.html",
	"category.html": "index.html",
	"author.html":   "index.html",
}

// HTML ..
func (s *Template) HTML(context map[string]interface{}) (string, error) {
	vars := G.GetContext()
	for k, v := range context {
		vars[k] = v
	}
	tpl := pongo2.Must(pongo2.FromFile(s.File()))
	return tpl.Execute(pongo2.Context(vars))
}

// Path ..
func (s *Template) File() string {
	file := filepath.Join(conf.Theme.Path, "templates", s.Name)
	if !FileExists(file) {
		if v, ok := templateRedirect[s.Name]; ok {
			file = filepath.Join(conf.Theme.Path, "templates", v)
		}
	}
	return file
}

// Exists ..
func (s *Template) Exists() bool {
	return FileExists(s.File())
}

// Format ..
func (s *Template) GetSavePath(context map[string]string) string {
	if context == nil {
		context = make(map[string]string)
	}
	if s.SavePath == "" {
		return s.Name
	}
	return StringReplace(s.SavePath, context)
}
