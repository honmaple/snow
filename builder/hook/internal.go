package hook

import (
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type internal struct {
	BaseHook
	conf config.Config
}

func (b *internal) Name() string {
	return "internal"
}

func (b *internal) BeforePagesWrite(pages page.Pages) page.Pages {
	if filter := b.conf.Get("page_filter"); filter != nil {
		pages = pages.Filter(filter)
	}
	if order := b.conf.GetString("page_orderby"); order != "" {
		pages = pages.OrderBy(order)
	}

	var (
		prev   *page.Page
		metas  = b.conf.GetStringMap("page_meta")
		labels = pages.GroupBy("type")
		npages = make(page.Pages, 0)
	)
	for _, label := range labels {
		if _, ok := metas[label.Name]; !ok {
			continue
		}

		var prevInType *page.Page
		for _, page := range label.List {
			page.PrevInType = prevInType
			if prevInType != nil {
				prevInType.NextInType = page
			}
			prevInType = page

			page.Prev = prev
			if prev != nil {
				prev.Next = page
			}
			prev = page
		}
		// 如果未写入详情页, 列表页也默认排除
		npages = append(npages, label.List...)
	}
	return npages
}

func init() {
	Register("internal", func(conf config.Config, theme theme.Theme) Hook {
		return &internal{conf: conf}
	})
}
