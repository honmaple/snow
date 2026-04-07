package types

type (
	Section struct {
		IsHome      bool
		FrontMatter *FrontMatter
		File        string

		Lang string

		Title       string
		Description string
		Summary     string
		Content     string

		Slug         string
		Path         string
		Permalink    string
		RelPermalink string

		Draft     bool
		Assets    []*Asset
		Pages     Pages
		WordCount int64
		Formats   Formats
	}
	Sections []*Section
)
