package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateHTMLPreservesRawHTMLWhenNotTruncated(t *testing.T) {
	input := `<p title='"quoted"' data-x="a&amp;b">Tom &amp; Jerry<br><img src="x" alt="a > b"></p>`

	assert.Equal(t, input, TruncateHTML(input, 10, "..."))
}

func TestTruncateHTMLUsesTextDataWhenTruncated(t *testing.T) {
	input := `<p>Tom &amp; Jerry walks</p>`

	assert.Equal(t, `<p>Tom & Jerry...</p>`, TruncateHTML(input, 2, "..."))
}
