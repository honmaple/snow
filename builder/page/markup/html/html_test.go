package html

import (
	"strings"
	"testing"

	"github.com/honmaple/snow/utils"
	"github.com/stretchr/testify/assert"
)

func assertFunc(t *testing.T, text string) {
	var (
		r = strings.NewReader(text)
	)

	meta, _ := readMeta(r)

	date, _ := utils.ParseTime("2023-02-24 20:35:51")
	assert.Equal(t, map[string]interface{}{
		"title":   "aaa",
		"date":    date,
		"tags":    []string{"snow", "hello", "world"},
		"authors": []string{"snow", "snow1"},
		"formats": map[string]interface{}{
			"atom": map[string]interface{}{
				"path":     "index.html",
				"template": "index.json",
			},
		},
		"custom_css": []string{"./main.css", "./body.css"},
		"custom_js":  []string{"./main.js"},
		"content":    "content",
	}, map[string]interface{}(meta))
}

func TestMeta(t *testing.T) {
	text := `<html>
  <head>
	<title>aaa</title>
	<meta name="date" content="2023-02-24 20:35:51" />
	<meta name="tags" content="[snow, hello, world]" />
	<meta name="authors" content="[snow, snow1]" />
	<meta name="formats.atom.path" content="index.html" />
	<meta name="formats.atom.template" content="index.json" />
	<link href="./main.css" rel="stylesheet">
	<link href="./body.css" rel="stylesheet">
	<script src="./main.js"></script>
	<script type="text/javascript">sss</script>
  </head>
  <body>
	content
  </body>
</html>
`
	assertFunc(t, text)
}
