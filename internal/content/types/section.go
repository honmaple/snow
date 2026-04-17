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
