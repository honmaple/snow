package markdown

import (
	"bytes"
	"strings"
	"testing"

	"github.com/honmaple/snow/utils"
	"github.com/stretchr/testify/assert"
)

func assertFunc(t *testing.T, text string) {
	var (
		r       = strings.NewReader(text)
		content bytes.Buffer
		summary bytes.Buffer
	)

	meta, _ := readMeta(r, &content, &summary)

	date, _ := utils.ParseTime("2023-02-24 20:35:51")
	assert.Equal(t, map[string]interface{}{
		"title":   "aaa",
		"date":    date,
		"tags":    []string{"snow", "hello, world"},
		"authors": []string{"snow", "snow1"},
		"formats": map[string]interface{}{
			"atom": map[string]interface{}{
				"path":     "index.html",
				"template": "index.json",
			},
		},
		"formats.js": map[string]interface{}{
			"atom": "index.xml",
		},
	}, map[string]interface{}(meta))

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
	assertFunc(t, text)
	assertFunc(t, text1)
}
