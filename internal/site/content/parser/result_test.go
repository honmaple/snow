package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetFrontMatterConvertsScalarValues(t *testing.T) {
	result := &Result{FrontMatter: make(map[string]any)}

	result.SetFrontMatter("Draft", "true")
	result.SetFrontMatter("Weight", "12")
	result.SetFrontMatter("Title", "Hello")

	assert.Equal(t, true, result.FrontMatter["draft"])
	assert.Equal(t, 12, result.FrontMatter["weight"])
	assert.Equal(t, "Hello", result.FrontMatter["title"])
}

func TestSetFrontMatterConvertsAndMergesSlices(t *testing.T) {
	result := &Result{FrontMatter: make(map[string]any)}

	result.SetFrontMatter("tags", "[go, snow]")
	result.SetFrontMatter("tags", "[parser]")

	assert.Equal(t, []any{"go", "snow", "parser"}, result.FrontMatter["tags"])
}

func TestSetFrontMatterBuildsNestedMaps(t *testing.T) {
	result := &Result{FrontMatter: make(map[string]any)}

	result.SetFrontMatter("formats.atom.path", "index.xml")
	result.SetFrontMatter("formats.atom.template", "atom.xml")
	result.SetFrontMatter("params.author.name", "Ada")

	assert.Equal(t, map[string]any{
		"atom": map[string]any{
			"path":     "index.xml",
			"template": "atom.xml",
		},
	}, result.FrontMatter["formats"])
	assert.Equal(t, map[string]any{
		"author": map[string]any{
			"name": "Ada",
		},
	}, result.FrontMatter["params"])
}
