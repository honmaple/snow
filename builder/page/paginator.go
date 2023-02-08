package page

import (
	"strconv"

	"github.com/honmaple/snow/utils"
)

type paginator struct {
	Next    *paginator
	Prev    *paginator
	Total   int
	PageNum int
	List    Pages
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

func (pages Pages) Paginator(number int, output string, vars map[string]string) []*paginator {
	var maxpage int

	length := len(pages)
	if number <= 0 {
		number = len(pages)
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
		pvars := map[string]string{
			"{number}":     numstr,
			"{number:one}": numstr,
		}
		for k, v := range vars {
			pvars[k] = v
		}
		if num == 0 {
			pvars["{number}"] = ""
		}
		por.URL = utils.StringReplace(output, pvars)

		end := (num + 1) * number
		if end > length {
			end = length
		}
		por.List = pages[num*number : end]
		pors[num] = por

		if prev != nil {
			prev.Next = por
		}
		prev = por
	}
	return pors
}
