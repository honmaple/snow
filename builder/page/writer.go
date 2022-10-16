package page

import (
	"fmt"
)

const (
	prefixTemplate   = "page_meta.%s.prefix"
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
		// 如果是已知类型，只写入详情页, 列表页由其他模板提供
		if ps, ok := types[name]; ok {
			if err := b.write(name, ps); err != nil {
				return err
			}
			continue
		}
		if err := b.writeOther(name, pages); err != nil {
			return err
		}
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

func (b *Builder) write(key string, pages Pages) error {
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		var prev *Page
		for i, page := range pages {
			page.Prev = prev
			if i < len(pages)-1 {
				page.Next = pages[i+1]
			}
			if err := b.theme.WriteTemplate(layouts, page.URL, map[string]interface{}{
				"page": b.hooks.BeforePage(page),
			}); err != nil {
				return err
			}
			prev = page
		}
	}
	return nil
}

func (b *Builder) writeOther(key string, pages Pages) error {
	var section Section

	listk := fmt.Sprintf("%s.list", key)

	if k := fmt.Sprintf(filterTemplate, listk); b.conf.IsSet(k) {
		pages = pages.Filter(b.conf.Get(k))
	}
	if by := b.conf.GetString(fmt.Sprintf(groupbyTemplate, listk)); by != "" {
		section = pages.GroupBy(by)
	} else {
		section = Section{"": pages}
	}
	if layouts, output := b.lookup(listk); len(layouts) > 0 && output != "" {
		paginate := b.conf.GetInt("page_paginate")
		if k := fmt.Sprintf(paginateTemplate, listk); b.conf.IsSet(k) {
			paginate = b.conf.GetInt(k)
		}

		section = b.hooks.BeforePageSection(section)
		for slug, pages := range section {
			pors := pages.Paginator(paginate, slug, output)
			for _, por := range pors {
				if err := b.theme.WriteTemplate(layouts, por.URL, map[string]interface{}{
					"slug":      slug,
					"pages":     por.List,
					"paginator": por,
				}); err != nil {
					return err
				}
			}
		}
	}
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		return b.theme.WriteTemplate(layouts, output, map[string]interface{}{
			"pages":   pages,
			"section": section,
		})
	}
	return nil
}
