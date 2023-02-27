package page

import (
	"path/filepath"
	"strconv"

	"github.com/honmaple/snow/utils"
)

type paginator struct {
	Next    *paginator
	Prev    *paginator
	Total   int
	PageNum int
	List    []interface{}
	URL     string
	All     []*paginator
}

func (p *paginator) Page(number int) *paginator {
	n := number - 1
	if n <= 0 {
		return p.All[0]
	}
	return p.All[n]
}

func (p *paginator) First() *paginator {
	return p.Page(1)
}

func (p *paginator) Last() *paginator {
	return p.Page(p.Total)
}

func (p *paginator) HasPrev() bool {
	return p.Prev != nil
}

func (p *paginator) HasNext() bool {
	return p.Next != nil
}

func Paginator(list []interface{}, number int, path string, paginatePath string) []*paginator {
	output := path
	if number > 0 {
		if paginatePath == "" {
			paginatePath = "{name}{number}{extension}"
		}
		file := filepath.Base(path)
		exts := filepath.Ext(file)
		output = filepath.Join(filepath.Dir(path), utils.StringReplace(paginatePath, map[string]string{
			"{name}":      file[:len(file)-len(exts)],
			"{extension}": exts,
		}))
	}

	var maxpage int

	length := len(list)
	if number <= 0 {
		number = len(list)
		maxpage = 1
	} else if length%number == 0 {
		maxpage = length / number
	} else {
		maxpage = length/number + 1
	}

	var prev *paginator

	pors := make([]*paginator, maxpage)
	for num := range pors {
		por := &paginator{
			Total:   maxpage,
			PageNum: num + 1,
			Prev:    prev,
			All:     pors,
		}
		numstr := strconv.Itoa(num + 1)
		vars := map[string]string{
			"{number}":     numstr,
			"{number:one}": numstr,
		}
		if num == 0 {
			vars["{number}"] = ""
		}
		por.URL = utils.StringReplace(output, vars)

		end := (num + 1) * number
		if end > length {
			end = length
		}
		por.List = list[num*number : end]
		pors[num] = por

		if prev != nil {
			prev.Next = por
		}
		prev = por
	}
	return pors
}
