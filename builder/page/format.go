package page

import (
	"github.com/spf13/viper"
)

type (
	Format struct {
		Name      string
		Path      string
		Template  string
		Permalink string
	}
	Formats []*Format
)

func (fs Formats) Find(name string) *Format {
	for _, f := range fs {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func (b *Builder) formats(meta Meta, realPath func(string) string) Formats {
	v := viper.New()
	v.MergeConfigMap(meta)

	formats := make(Formats, 0)
	for name := range meta.GetStringMap("formats") {
		path := v.GetString("formats." + name + ".path")
		template := v.GetString("formats." + name + ".template")
		if template == "" {
			template = b.conf.GetString("formats." + name + ".template")
		}
		if path == "" || template == "" {
			continue
		}

		if realPath != nil {
			path, template = realPath(path), realPath(template)
		}
		format := &Format{
			Name:     name,
			Template: template,
		}
		format.Path = b.conf.GetRelURL(path)
		format.Permalink = b.conf.GetURL(format.Path)

		formats = append(formats, format)
	}
	return formats
}
