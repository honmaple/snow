package markdown

import (
	"strings"
	"testing"

	"github.com/honmaple/snow/internal/utils"
	"github.com/stretchr/testify/assert"
)

func assertFunc(t *testing.T, text string) {
	r := New(&Option{})

	result, err := r.Parse(strings.NewReader(text))
	assert.Nil(t, err)

	date, _ := utils.ParseTime("2023-02-24 20:35:51")

	expected := map[string]any{
		"title":   "aaa",
		"date":    date,
		"tags":    []any{"snow", "hello, world"},
		"authors": []any{"snow", "snow1"},
		"formats": map[string]any{
			"atom": map[string]any{
				"path":     "index.html",
				"template": "index.json",
			},
			"js": map[string]any{
				"atom": "index.xml",
			},
		},
	}
	for k, v := range result.FrontMatter {
		assert.Equal(t, expected[k], v)
	}
	assert.Equal(t, "\nsummary\n", result.RawSummary)
	assert.Equal(t, "\nsummary\n<!--more-->\ncontent\n", result.RawContent)
}

func TestMeta(t *testing.T) {
	text := `---
title: aaa
date: 2023-02-24 20:35:51
tags:
  - snow
  - hello, world
authors:
  - snow
  - snow1
formats:
  atom:
   path: index.html
   template: index.json
"formats.js":
  atom: index.xml
---

summary
<!--more-->
content
`
//	text1 := `Title:   aaa
// Date:     2023-02-24 20:35:51
// Tags:     [snow,"hello, world"]
// Authors:  [snow]
// Authors:  [snow1]
// Formats.atom.path: index.html
// Formats.atom.template: index.json
// "formats.js".atom: index.xml

// summary
// <!--more-->
// content
// `

//	text2 := `+++
// title =   "aaa"
// date =    "2023-02-24 20:35:51"
// tags =     ["snow","hello, world"]
// authors =  ["snow", "snow1"]
// formats.atom.path = "index.html"
// formats.atom.template = "index.json"
// ["formats.js"]
//   atom = "index.xml"
// +++

// summary
// <!--more-->
// content
// `
	assertFunc(t, text)
	// assertFunc(t, text1)
	// assertFunc(t, text2)
}
