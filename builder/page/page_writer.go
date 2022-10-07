package page

import (
	"fmt"
	// "io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flosch/pongo2/v4"
	"github.com/honmaple/snow/utils"
)

const (
	ignoreTemplate   = "theme.templates.%s.ignore"
	lookupTemplate   = "theme.templates.%s.lookup"
	outputTemplate   = "theme.templates.%s.output"
	filterTemplate   = "theme.templates.%s.filter"
	groupbyTemplate  = "theme.templates.%s.groupby"
	paginateTemplate = "theme.templates.%s.paginate"
)

func (b *Builder) Write(pages Pages) error {
	templates := b.conf.GetStringMap("theme.templates")
	fmt.Println(templates)

	types := pages.GroupBy("type")
	for name := range templates {
		// 如果是已知类型，只写入详情页, 列表页由其他模板提供
		if b.types[name] {
			b.write(name, types)
			continue
		}
		b.writeOther(name, pages)
	}
	return nil
}

func (b *Builder) lookup(key string) (string, string) {
	output := b.conf.GetString(fmt.Sprintf(outputTemplate, key))
	if output == "" {
		return "", ""
	}

	names := b.conf.GetStringSlice(fmt.Sprintf(lookupTemplate, key))
	if len(names) == 0 {
		return "", ""
	}

	layouts := b.conf.GetString("layouts_dir")
	for _, name := range names {
		file := filepath.Join(layouts, name)
		if utils.FileExists(file) {
			return file, output
		}
	}

	layout := filepath.Join(b.conf.GetString("theme.path"), "templates")
	for _, name := range names {
		file := filepath.Join(layout, name)
		if utils.FileExists(file) {
			return file, output
		}
	}
	return "", ""
}

func (b *Builder) write(key string, section Section) {
	pages := section[Label{Name: key}]
	for label, pages := range section {
		fmt.Println(label.Name, key, len(pages))
	}

	if layout, output := b.lookup(key); layout != "" && output != "" {
		for _, page := range pages {
			b.writeTemplate(layout, page.URL, map[string]interface{}{
				"page": page,
			})
		}
	}
}

func (b *Builder) writeOther(key string, pages Pages) {
	var section Section

	listk := fmt.Sprintf("%s.list", key)

	if k := fmt.Sprintf(filterTemplate, listk); b.conf.IsSet(k) {
		pages = pages.Filter(b.conf.GetStringMap(k))
	}
	if by := b.conf.GetString(fmt.Sprintf(groupbyTemplate, listk)); by != "" {
		section = pages.GroupBy(by)
	} else {
		section = Section{Label{}: pages}
	}
	if layout, output := b.lookup(listk); layout != "" && output != "" {
		paginate := b.conf.GetInt("theme.paginate")
		if k := fmt.Sprintf(paginateTemplate, listk); b.conf.IsSet(k) {
			paginate = b.conf.GetInt(k)
		}

		newSection := make(Section)
		for label, pages := range section {
			for index, por := range pages.Paginator(paginate) {
				num := strconv.Itoa(index + 1)
				vars := map[string]string{
					"{slug}":       label.Name,
					"{number}":     num,
					"{number:one}": num,
				}
				if index == 0 {
					vars["{number}"] = ""
				}
				dest := utils.StringReplace(output, vars)
				b.writeTemplate(layout, dest, map[string]interface{}{
					"slug":      label.Name,
					"pages":     pages,
					"paginator": por,
				})
			}
			newLabel := Label{
				URL: utils.StringReplace(output, map[string]string{
					"{slug}":       label.Name,
					"{number}":     "",
					"{number:one}": "1",
				}),
				Name: label.Name,
			}
			newSection[newLabel] = pages
		}
		section = newSection
	}
	if layout, output := b.lookup(key); layout != "" && output != "" {
		b.writeTemplate(layout, output, map[string]interface{}{
			"pages":   pages,
			"section": section,
		})
	}
}

func (b *Builder) writeFile(file, content string) error {
	fmt.Println(file)
	return nil
	// writefile := filepath.Join(b.conf.GetString("output_dir"), file)
	// if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
	// 	os.MkdirAll(dir, 0755)
	// }
	// return ioutil.WriteFile(writefile, []byte(content), 0755)
}

func (b *Builder) writeTemplate(tmpl string, file string, context map[string]interface{}) error {
	if file == "" {
		return nil
	}
	writefile := filepath.Join(b.conf.GetString("output_dir"), file)
	if dir := filepath.Dir(writefile); !utils.FileExists(dir) {
		os.MkdirAll(dir, 0755)
	}

	tpl := pongo2.Must(pongo2.FromFile(tmpl))
	f, err := os.OpenFile(writefile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	c := make(map[string]interface{})
	for k, v := range b.context {
		c[k] = v
	}
	for k, v := range context {
		c[k] = v
	}
	fmt.Println("write file to: ", writefile)
	return tpl.ExecuteWriter(c, f)
}
