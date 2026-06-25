package content

import (
	"sort"
	"strings"

	"github.com/honmaple/snow/internal/utils"
)

type (
	PageGroup struct {
		Name  string
		Pages Pages

		Parent   *PageGroup
		Children PageGroups
	}
	PageGroups []*PageGroup
)

func SortPageGroups(groups PageGroups, key string, recursive bool) {
	if key == "" {
		key = "name"
	}
	sort.SliceStable(groups, utils.Sort(key, func(k string, i int, j int) int {
		switch k {
		case "-":
			return 0 - strings.Compare(groups[i].Name, groups[j].Name)
		case "name":
			return strings.Compare(groups[i].Name, groups[j].Name)
		case "count":
			return utils.Compare(len(groups[i].Pages), len(groups[j].Pages))
		default:
			return 0
		}
	}))
	if recursive {
		for _, group := range groups {
			SortPageGroups(group.Children, key, true)
		}
	}
}

func (groups PageGroups) Reverse() PageGroups {
	ns := make(PageGroups, len(groups))
	for i, j := 0, len(groups)-1; j >= 0; i, j = i+1, j-1 {
		ns[i] = groups[j]
	}
	return ns
}

func (groups PageGroups) OrderBy(key string) PageGroups {
	newGroups := make(PageGroups, len(groups))
	copy(newGroups, groups)

	SortPageGroups(newGroups, key, true)
	return newGroups
}
