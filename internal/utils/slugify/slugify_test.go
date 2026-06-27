package slugify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeDefault(t *testing.T) {
	assert.Equal(t, "hello-world-2026", Make("Hello, World 2026!"))
	assert.Equal(t, "docs-v1", Make("Docs.v1"))
	assert.Equal(t, "ying-shi", Make("影師"))
}

func TestMakePreserveUnicode(t *testing.T) {
	assert.Equal(t, "你好-world", Make("你好 World", WithPreserveUnicode(true)))
}

func TestMakePreserveChars(t *testing.T) {
	assert.Equal(t, "docs.v1", Make("Docs.v1", WithPreserveChars(".")))
}
