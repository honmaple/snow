package content

import (
	stdpath "path"
	"strconv"
	"strings"
)

type (
	Pager struct {
		Path      string
		Permalink string
		Pages     Pages
		PageNum   int
		Next      *Pager
		Prev      *Pager
	}
	Pagers []*Pager
)

type Paginator struct {
	*Pager
	Total  int
	Pagers Pagers
}

func (p *Paginator) First() *Pager {
	return p.Page(1)
}

func (p *Paginator) Last() *Pager {
	return p.Page(p.Total)
}

func (p *Paginator) HasPrev() bool {
	return p.Prev != nil
}

func (p *Paginator) HasNext() bool {
	return p.Next != nil
}

func (p *Paginator) Page(number int) *Pager {
	if p == nil || len(p.Pagers) == 0 {
		return nil
	}
	n := number - 1
	if n <= 0 {
		return p.Pagers[0]
	}
	if n >= len(p.Pagers) {
		return p.Pagers[len(p.Pagers)-1]
	}
	return p.Pagers[n]
}

func (d *Processor) resolvePaginatePath(num int, basePath string, paginatePath string) string {
	if num <= 1 {
		return d.resolvePath(basePath, nil)
	}
	if paginatePath == "" {
		if strings.HasSuffix(basePath, "/") {
			paginatePath = "page/{number}/"
		} else {
			paginatePath = "{name}{number}{extension}"
		}
	}
	vars := map[string]string{
		"{number}": strconv.Itoa(num),
	}
	if strings.HasSuffix(basePath, "/") {
		vars["{name}"] = "index"
		vars["{extension}"] = ".html"
		return d.resolvePath(basePath+paginatePath, vars)
	}

	ext := stdpath.Ext(basePath)
	dir, name := stdpath.Split(basePath)

	vars["{name}"] = strings.TrimSuffix(name, ext)
	vars["{extension}"] = ext
	return d.resolvePath(stdpath.Join(dir, paginatePath), vars)
}

func (d *Processor) PaginateBy(pages Pages, size int, basePath string, paginatePath string, lang string) Pagers {
	var (
		total  = 1
		length = len(pages)
	)
	if size <= 0 {
		size = length
		total = 1
	} else if length == 0 {
		total = 1
	} else if length%size == 0 {
		total = length / size
	} else {
		total = length/size + 1
	}

	var (
		prev *Pager
		pors = make(Pagers, total)
		lctx = d.ctx.For(lang)
	)
	for num := 1; num <= total; num++ {
		por := &Pager{
			PageNum: num,
			Prev:    prev,
		}
		por.Path = d.resolvePaginatePath(num, basePath, paginatePath)
		por.Permalink = lctx.GetURL(por.Path)

		end := num * size
		if end > length {
			end = length
		}
		start := (num - 1) * size
		if start > length {
			start = length
		}
		por.Pages = pages[start:end]

		if prev != nil {
			prev.Next = por
		}
		prev = por

		pors[num-1] = por
	}
	return pors
}

func NewPaginator(pager *Pager, pagers Pagers) *Paginator {
	return &Paginator{
		Total:  len(pagers),
		Pager:  pager,
		Pagers: pagers,
	}
}
