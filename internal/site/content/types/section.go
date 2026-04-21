package types

type (
	Section struct {
		*Node

		Path      string
		Permalink string

		Pages    Pages
		Assets   Assets
		Formats  Formats
		Children Sections
	}
	Sections []*Section
)

func (secs Sections) Len() int           { return len(secs) }
func (secs Sections) Swap(i, j int)      { secs[i], secs[j] = secs[j], secs[i] }
func (secs Sections) Less(i, j int) bool { return secs[i].Title < secs[j].Title }
