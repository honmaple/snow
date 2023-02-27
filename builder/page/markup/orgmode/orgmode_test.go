package orgmode

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
	assert.Equal(t, "\nsummary\n#+MORE\ncontent\n", content.String())
}

func TestMeta(t *testing.T) {
	text := `:PROPERTIES:
:title:   aaa
:date:     2023-02-24 20:35:51
:tags:     [snow,"hello, world"]
:authors:  [snow]
:authors:  [snow1]
:formats.atom.path: index.html
:formats.atom.template: index.json
:"formats.js".atom: index.xml
:END:

summary
#+MORE
content
`
	text1 := `#+TITLE: aaa
#+DATE: 2023-02-24 20:35:51
#+PROPERTY: TAGS [snow,"hello, world"]
#+PROPERTY: AUTHORS [snow]
#+PROPERTY: AUTHORS [snow1]
#+PROPERTY: formats.atom.path     index.html
#+PROPERTY: formats.atom.template index.json
#+PROPERTY: "formats.js".atom     index.xml

summary
#+MORE
content
`
	assertFunc(t, text)
	assertFunc(t, text1)
}
