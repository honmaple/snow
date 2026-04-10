package types

type (
	Section struct {
		File        *File
		FrontMatter *FrontMatter

		Lang string

		Title       string
		Description string
		Summary     string
		Content     string
		RawContent  string

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
