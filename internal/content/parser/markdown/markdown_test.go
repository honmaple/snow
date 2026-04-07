package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/honmaple/snow/internal/content/parser"
	"github.com/honmaple/snow/internal/utils"
	"github.com/stretchr/testify/assert"
)

func assertFunc(t *testing.T, text string) {
	var (
		r       = strings.NewReader(text)
		content bytes.Buffer
		summary bytes.Buffer
	)

	result := &parser.Result{
		FrontMatter: make(map[string]any),
	}

	_ = readMeta(r, &content, &summary, result)

	date, _ := utils.ParseTime("2023-02-24 20:35:51")
	assert.Equal(t, map[string]any{
		"title":   "aaa",
		"date":    date,
		"tags":    []string{"snow", "hello, world"},
		"authors": []string{"snow", "snow1"},
		"formats": map[string]any{
			"atom": map[string]any{
				"path":     "index.html",
				"template": "index.json",
			},
		},
		"formats.js": map[string]any{
			"atom": "index.xml",
		},
	}, result.FrontMatter)

	assert.Equal(t, "\nsummary\n", summary.String())
	assert.Equal(t, "\nsummary\n<!--more-->\ncontent\n", content.String())
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
	text1 := `Title:   aaa
Date:     2023-02-24 20:35:51
Tags:     [snow,"hello, world"]
Authors:  [snow]
Authors:  [snow1]
Formats.atom.path: index.html
Formats.atom.template: index.json
"formats.js".atom: index.xml

summary
<!--more-->
content
`

	text2 := `+++
title =   "aaa"
date =    "2023-02-24 20:35:51"
tags =     ["snow","hello, world"]
authors =  ["snow", "snow1"]
formats.atom.path = "index.html"
formats.atom.template = "index.json"
["formats.js"]
  atom = "index.xml"
+++

summary
<!--more-->
content
`
	assertFunc(t, text)
	assertFunc(t, text1)
	assertFunc(t, text2)
}
