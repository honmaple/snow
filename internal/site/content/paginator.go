package content

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/honmaple/snow/internal/utils"
)

type Paginator[T any] struct {
	Next    *Paginator[T]
	Prev    *Paginator[T]
	Total   int
	PageNum int
	List    []T
	Path    string
	All     []*Paginator[T]
}

func (p *Paginator[T]) Page(number int) *Paginator[T] {
	n := number - 1
	if n <= 0 {
		return p.All[0]
	}
	return p.All[n]
}

func (p *Paginator[T]) First() *Paginator[T] {
	return p.Page(1)
}

func (p *Paginator[T]) Last() *Paginator[T] {
	return p.Page(p.Total)
}

func (p *Paginator[T]) HasPrev() bool {
	return p.Prev != nil
}

func (p *Paginator[T]) HasNext() bool {
	return p.Next != nil
}

func Paginate[T any](list []T, number int, path string, paginatePath string) []*Paginator[T] {
	output := path
	if number > 0 {
		if paginatePath == "" {
			paginatePath = "{name}{number}{extension}"
		}
		name, exts := "", ".html"
		if !strings.HasSuffix(path, "/") {
			file := filepath.Base(path)
			exts = filepath.Ext(file)
			name = file[:len(file)-len(exts)]
		}
		output = filepath.Join(filepath.Dir(path), utils.StringReplace(paginatePath, map[string]string{
			"{name}":      name,
			"{extension}": exts,
		}))
	}

	var maxpage int

	length := len(list)
	if length == 0 {
		number = 0
		maxpage = 1
	} else if number <= 0 {
		number = len(list)
		maxpage = 1
	} else if length%number == 0 {
		maxpage = length / number
	} else {
		maxpage = length/number + 1
	}

	var prev *Paginator[T]

	pors := make([]*Paginator[T], maxpage)
	for num := range pors {
		por := &Paginator[T]{
			Total:   maxpage,
			PageNum: num + 1,
			Prev:    prev,
			All:     pors,
		}
		if num == 0 {
			por.Path = path
		} else {
			por.Path = utils.StringReplace(output, map[string]string{
				"{number}": strconv.Itoa(num + 1),
			})
		}

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
