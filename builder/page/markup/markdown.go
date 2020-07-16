package markup

func (m *Markup) markdown(file string) (map[string]string, error) {
	meta := map[string]string{
		"title": "markdown",
		// Title:    metadata.Title,
		// Summary:  metadata.Summary,
		// Content:  metadata.Content,
		// Date:     metadata.Date,
		// Tags:     metadata.Tags,
		// Category: metadata.Category,
		// Author:   metadata.Author,
		// Authors:  metadata.Authors,
	}
	return meta, nil
}
