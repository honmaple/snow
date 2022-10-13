package page

import (
	"fmt"
	// "io/ioutil"
	"strconv"

	"github.com/honmaple/snow/utils"
)

const (
	ignoreTemplate   = "page_meta.%s.ignore"
	lookupTemplate   = "page_meta.%s.lookup"
	outputTemplate   = "page_meta.%s.output"
	filterTemplate   = "page_meta.%s.filter"
	groupbyTemplate  = "page_meta.%s.groupby"
	paginateTemplate = "page_meta.%s.paginate"
)

func (b *Builder) Write(pages Pages) error {
	metas := b.conf.GetStringMap("page_meta")

	types := pages.GroupBy("type")
	for name := range metas {
		// fmt.Println(name)
		// 如果是已知类型，只写入详情页, 列表页由其他模板提供
		if b.types[name] {
			b.write(name, types)
			continue
		}
		b.writeOther(name, pages)
	}
	return nil
}

func (b *Builder) lookup(key string) ([]string, string) {
	output := b.conf.GetString(fmt.Sprintf(outputTemplate, key))
	if output == "" {
		return nil, ""
	}

	names := b.conf.GetStringSlice(fmt.Sprintf(lookupTemplate, key))
	return names, output
}

func (b *Builder) write(key string, section Section) {
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		pages := section[Label{Name: key}]

		var prev *Page
		for i, page := range pages {
			page.Prev = prev
			if i < len(pages)-1 {
				page.Next = pages[i+1]
			}
			b.theme.WriteTemplate(layouts, page.URL, map[string]interface{}{
				"page": b.hooks.BeforePage(page),
			})
			prev = page
		}
	}
}

func (b *Builder) writeOther(key string, pages Pages) {
	var section Section

	listk := fmt.Sprintf("%s.list", key)

	if k := fmt.Sprintf(filterTemplate, listk); b.conf.IsSet(k) {
		pages = pages.Filter(b.conf.Get(k))
	}
	if by := b.conf.GetString(fmt.Sprintf(groupbyTemplate, listk)); by != "" {
		section = pages.GroupBy(by)
	} else {
		section = Section{Label{}: pages}
	}
	if layouts, output := b.lookup(listk); len(layouts) > 0 && output != "" {
		paginate := b.conf.GetInt("page_paginate")
		if k := fmt.Sprintf(paginateTemplate, listk); b.conf.IsSet(k) {
			paginate = b.conf.GetInt(k)
		}

		newSection := make(Section)
		section = b.hooks.BeforePageSection(section)
		for label, pages := range section {
			for index, por := range pages.Paginator(paginate) {
				// if label.Name == "" {
				//	continue
				// }
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
				b.theme.WriteTemplate(layouts, dest, map[string]interface{}{
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
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		b.theme.WriteTemplate(layouts, output, map[string]interface{}{
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
	//	os.MkdirAll(dir, 0755)
	// }
	// return ioutil.WriteFile(writefile, []byte(content), 0755)
}
