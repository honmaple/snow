package page

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeta(t *testing.T) {
	m := make(Meta)
	m.Set("a", "b")
	m.Set("a", "c")
	assert.Equal(t, "c", m.Get("a"))

	m.Set("a.b", "b")
	m.Set("a.c", "c")
	assert.Equal(t, map[string]interface{}{
		"b": "b",
		"c": "c",
	}, m.Get("a"))
	m.Set("a.c.d", "c")
	assert.Equal(t, map[string]interface{}{
		"b": "b",
		"c": map[string]interface{}{
			"d": "c",
		},
	}, m.Get("a"))

	m = make(Meta)
	m.Set("a.b", "[b]")
	m.Set("a.b", `[c, "d,e"]`)
	assert.Equal(t, map[string]interface{}{
		"b": []interface{}{"b", "c", "d,e"},
	}, m.Get("a"))

	m = make(Meta)
	m.Set("a.a", "false")
	m.Set("a.b", `"false"`)
	m.Set("a.d", "True")
	assert.Equal(t, map[string]interface{}{
		"a": false,
		"b": `"false"`,
		"d": true,
	}, m.Get("a"))

	m = make(Meta)
	m.Set("a.a", "1")
	m.Set("a.b", "12")
	assert.Equal(t, map[string]interface{}{
		"a": 1,
		"b": 12,
	}, m.Get("a"))
}
