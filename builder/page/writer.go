package page

import (
	"fmt"
	"strings"
)

const (
	lookupTemplate   = "page_meta.%s.lookup"
	outputTemplate   = "page_meta.%s.output"
	filterTemplate   = "page_meta.%s.filter"
	orderbyTemplate  = "page_meta.%s.orderby"
	groupbyTemplate  = "page_meta.%s.groupby"
	paginateTemplate = "page_meta.%s.paginate"
)

func (b *Builder) Write(pages Pages, labels Labels) error {
	var (
		types = make(map[string]bool)
		metas = b.conf.GetStringMap("page_meta")
	)
	for _, label := range labels {
		types[label.Name] = true
		// 如果是已知类型，只写入详情页, 列表页由其他模板提供
		if err := b.write(label.Name, label.List); err != nil {
			return err
		}
	}

	// 写入列表页或者归档页
	for name := range metas {
		if types[name] {
			continue
		}
		if err := b.writeSingle(name, pages); err != nil {
			return err
		}
		if err := b.writeList(name, pages); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) list(key string, pages Pages) Pages {
	if k := fmt.Sprintf(filterTemplate, key); b.conf.IsSet(k) {
		pages = pages.Filter(b.conf.Get(k))
	}

	if k := fmt.Sprintf(orderbyTemplate, key); b.conf.IsSet(k) {
		pages = pages.OrderBy(b.conf.GetString(k))
	}
	return pages
}

func (b *Builder) lookup(key string) ([]string, string) {
	output := b.conf.GetString(fmt.Sprintf(outputTemplate, key))
	if output == "" {
		return nil, ""
	}

	layouts := b.conf.GetStringSlice(fmt.Sprintf(lookupTemplate, key))
	return layouts, output
}

func (b *Builder) write(key string, pages Pages) error {
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		for _, page := range b.list(key, pages) {
			if err := b.theme.WriteTemplate(layouts, page.URL, map[string]interface{}{
				"page": b.hooks.BeforePageWrite(page),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Builder) writeSingle(key string, pages Pages) error {
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		pages = b.list(key, pages)

		vars := map[string]interface{}{
			"pages": pages,
		}
		if by := b.conf.GetString(fmt.Sprintf(groupbyTemplate, key)); by != "" {
			vars["labels"] = pages.GroupBy(by)
		}
		return b.theme.WriteTemplate(layouts, output, vars)
	}
	return nil
}

func (b *Builder) writeList(key string, pages Pages) error {
	key = fmt.Sprintf("%s.list", key)
	if layouts, output := b.lookup(key); len(layouts) > 0 && output != "" {
		paginate := b.conf.GetInt("page_paginate")
		if k := fmt.Sprintf(paginateTemplate, key); b.conf.IsSet(k) {
			paginate = b.conf.GetInt(k)
		}

		pages = b.list(key, pages)

		var labels Labels
		if by := b.conf.GetString(fmt.Sprintf(groupbyTemplate, key)); by != "" {
			labels = pages.GroupBy(by)
		} else {
			labels = Labels{&Label{List: pages}}
		}

		labels = b.hooks.BeforeLabelsWrite(labels)
		if err := b.writeLabels(labels, paginate, layouts, output); err != nil {
			return err
		}
	}
	return nil
}

func (b *Builder) writeLabels(labels Labels, paginate int, layouts []string, output string) error {
	for _, label := range labels {
		pors := label.List.Paginator(paginate, strings.ToLower(label.Name), output)
		for _, por := range pors {
			if err := b.theme.WriteTemplate(layouts, por.URL, map[string]interface{}{
				"slug":      label.Name,
				"pages":     por.List,
				"paginator": por,
			}); err != nil {
				return err
			}
		}
		if len(label.Children) == 0 {
			continue
		}
		if err := b.writeLabels(label.Children, paginate, layouts, output); err != nil {
			return err
		}
	}
	return nil
}
