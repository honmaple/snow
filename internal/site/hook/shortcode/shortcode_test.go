package shortcode

import (
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTemplate struct {
	execute func(map[string]any) (string, error)
}

func (tpl *testTemplate) Name() string {
	return "test"
}

func (tpl *testTemplate) Execute(vars map[string]any) (string, error) {
	return tpl.execute(vars)
}

func testShortcodeSet(tpls map[string]*testTemplate) *ShortcodeSet {
	set := &ShortcodeSet{
		ctx:  &core.Context{},
		tpls: make(map[string]template.Template, len(tpls)),
	}
	for name, tpl := range tpls {
		set.tpls[name] = tpl
	}
	return set
}

func testPage() *content.Page {
	return &content.Page{
		Node: &content.Node{
			File: &content.File{Path: "content/test.md"},
		},
	}
}

func TestRenderShortcodeUsesBooleanAttributeAsName(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"youtube": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "youtube", vars["name"])
				assert.Equal(t, "youtube", vars["_name"])
				assert.Equal(t, 0, vars["counter"])
				assert.Equal(t, 0, vars["_counter"])

				params, ok := vars["params"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "xxx", params["id"])
				assert.NotContains(t, params, "youtube")
				return `<iframe data-id="` + params["id"].(string) + `"></iframe>`, nil
			},
		},
	})

	result := set.Render(testPage(), `<p>before</p><shortcode youtube id="xxx" />`)

	assert.Equal(t, `<p>before</p><iframe data-id="xxx"></iframe>`, result)
}

func TestRenderShortcodeBodyWithBooleanAttributeName(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"code": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "code", vars["name"])
				assert.Equal(t, "hello", vars["body"])

				params, ok := vars["params"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "go", params["lang"])
				assert.NotContains(t, params, "code")
				return `<pre data-lang="` + params["lang"].(string) + `">` + vars["body"].(string) + `</pre>`, nil
			},
		},
	})

	result := set.Render(testPage(), `<shortcode code lang="go">hello</shortcode>`)

	assert.Equal(t, `<pre data-lang="go">hello</pre>`, result)
}

func TestRenderShortcodeKeepsLegacyNameAttribute(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"youtube": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "youtube", vars["name"])

				params, ok := vars["params"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "xxx", params["id"])
				assert.NotContains(t, params, "_name")
				return "legacy:" + params["id"].(string), nil
			},
		},
	})

	result := set.Render(testPage(), `<shortcode _name="youtube" id="xxx" />`)

	assert.Equal(t, `legacy:xxx`, result)
}
