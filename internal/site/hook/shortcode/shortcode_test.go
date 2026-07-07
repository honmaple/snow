package shortcode

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/template"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
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

func testSection() *content.Section {
	return &content.Section{
		Node: &content.Node{
			File: &content.File{Path: "content/_index.md"},
			Lang: "en",
		},
	}
}

func captureShortcodeWarnings(set *ShortcodeSet) *bytes.Buffer {
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	})
	set.ctx = &core.Context{Logger: logger}
	return &buf
}

func TestRenderShortcodeUsesBooleanAttributeAsName(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"youtube": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "youtube", vars["name"])
				assert.Equal(t, "youtube", vars["_name"])
				assert.Equal(t, 0, vars["counter"])
				assert.Equal(t, 0, vars["_counter"])

				assert.Equal(t, Params{"id": "xxx"}, vars["params"])
				return `<iframe data-id="xxx"></iframe>`, nil
			},
		},
	})

	result := set.Render("content/test.md", `<p>before</p><shortcode youtube id="xxx" />`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<p>before</p><iframe data-id="xxx"></iframe>`, result)
}

func TestRenderShortcodeBodyWithBooleanAttributeName(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"code": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "code", vars["name"])
				assert.Equal(t, "hello", vars["body"])

				assert.Equal(t, Params{"lang": "go"}, vars["params"])
				return `<pre data-lang="go">hello</pre>`, nil
			},
		},
	})

	result := set.Render("content/test.md", `<shortcode code lang="go">hello</shortcode>`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<pre data-lang="go">hello</pre>`, result)
}

func TestRenderShortcodeBodyWithDirectTagName(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"code": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "code", vars["name"])
				assert.Equal(t, "hello", vars["body"])
				assert.Equal(t, Params{"lang": "go"}, vars["params"])
				return `<pre data-lang="go">hello</pre>`, nil
			},
		},
	})

	result := set.Render("content/test.md", `<code lang="go">hello</code>`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<pre data-lang="go">hello</pre>`, result)
}

func TestRenderShortcodeBodyCanUnmarshalYAML(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"projects": {
			execute: func(vars map[string]any) (string, error) {
				var projects []map[string]any
				body := vars["body"].(string)
				if !assert.NoError(t, yaml.Unmarshal([]byte(body), &projects)) {
					return "", nil
				}
				if !assert.Len(t, projects, 2) {
					return "", nil
				}
				assert.Equal(t, "Snow", projects[0]["name"])
				assert.Equal(t, []any{"static-site", "go", "writing"}, projects[0]["tags"])
				assert.Equal(t, "Drift", projects[1]["name"])
				return "projects", nil
			},
		},
	})

	result := set.Render("content/test.md", `<shortcode projects>
- name: "Snow"
  description: "xxx"
  tags:
    - "static-site"
    - "go"
    - "writing"
- name: "Drift"
  description: "xxxx"
  tags:
    - "theme"
    - "tailwindcss"
</shortcode>`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `projects`, result)
}

func TestRenderShortcodeKeepsLegacyNameAttribute(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"youtube": {
			execute: func(vars map[string]any) (string, error) {
				assert.Equal(t, "youtube", vars["name"])

				assert.Equal(t, Params{"id": "xxx"}, vars["params"])
				return "legacy:xxx", nil
			},
		},
	})

	result := set.Render("content/test.md", `<shortcode _name="youtube" id="xxx" />`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `legacy:xxx`, result)
}

func TestParamsString(t *testing.T) {
	params := Params{
		"id":       "a&b",
		"disabled": "",
		"class":    "video",
	}

	assert.Equal(t, `class="video" disabled id="a&amp;b"`, params.String())
}

func TestParamsPop(t *testing.T) {
	params := Params{
		"id":    "xxx",
		"class": "video",
	}

	assert.Equal(t, "xxx", params.Pop("id"))
	assert.Equal(t, Params{"class": "video"}, params)
	assert.Nil(t, params.Pop("missing"))
}

func TestParamsGet(t *testing.T) {
	params := Params{"id": "xxx"}

	assert.Equal(t, "xxx", params.Get("id"))
	assert.Nil(t, params.Get("missing"))
}

func TestRenderShortcodeCountersUseSourceState(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"item": {
			execute: func(vars map[string]any) (string, error) {
				return fmt.Sprintf("item:%d", vars["counter"]), nil
			},
		},
		"wrap": {
			execute: func(vars map[string]any) (string, error) {
				return fmt.Sprintf("wrap:%d[%s]", vars["counter"], vars["body"]), nil
			},
		},
	})

	result := set.Render("content/test.md", `<shortcode item /><shortcode wrap><shortcode item /></shortcode><shortcode item />`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `item:0wrap:0[item:1]item:2`, result)
}

func TestRenderShortcodePreservesNonShortcodeHTML(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"ok": {
			execute: func(vars map[string]any) (string, error) {
				return "OK", nil
			},
		},
	})

	input := `<p title='"quoted"' data-x="a&amp;b">Tom &amp; Jerry<br><img src="x" alt="a > b"></p><shortcode ok />`
	result := set.Render("content/test.md", input, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<p title='"quoted"' data-x="a&amp;b">Tom &amp; Jerry<br><img src="x" alt="a > b"></p>OK`, result)
}

func TestRenderShortcodeBreaksAndWarnsOnExecuteError(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"broken": {
			execute: func(vars map[string]any) (string, error) {
				return "", errors.New("filter failed")
			},
		},
	})
	logs := captureShortcodeWarnings(set)

	result := set.Render("content/test.md", `<shortcode broken />`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<shortcode broken />`, result)
	assert.Contains(t, logs.String(), "filter failed")
}

func TestRenderShortcodeFallsBackWithBodyOnExecuteError(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"broken": {
			execute: func(vars map[string]any) (string, error) {
				return "", errors.New("filter failed")
			},
		},
	})

	result := set.Render("content/test.md", `<shortcode broken>body</shortcode>`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<shortcode broken>body</shortcode>`, result)
}

func TestRenderShortcodeFallsBackNestedBodyOnExecuteError(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"broken": {
			execute: func(vars map[string]any) (string, error) {
				return "", errors.New("filter failed")
			},
		},
		"ok": {
			execute: func(vars map[string]any) (string, error) {
				return "OK", nil
			},
		},
	})

	result := set.Render("content/test.md", `<shortcode broken>before <shortcode ok /> after</shortcode>`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<shortcode broken>before OK after</shortcode>`, result)
}

func TestRenderShortcodeWarnsAndKeepsUnclosedBody(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"notice": {
			execute: func(vars map[string]any) (string, error) {
				return "ignored", nil
			},
		},
	})
	logs := captureShortcodeWarnings(set)

	result := set.Render("content/test.md", `<shortcode notice>before <shortcode ok /> after`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<shortcode notice>before <shortcode ok /> after`, result)
	assert.Contains(t, logs.String(), `closing delimiter '</shortcode>' is missing`)
}

func TestRenderShortcodeWarnsAndKeepsUnclosedNestedBody(t *testing.T) {
	set := testShortcodeSet(map[string]*testTemplate{
		"wrap": {
			execute: func(vars map[string]any) (string, error) {
				return "wrap", nil
			},
		},
		"ok": {
			execute: func(vars map[string]any) (string, error) {
				return "OK", nil
			},
		},
	})
	logs := captureShortcodeWarnings(set)

	result := set.Render("content/test.md", `<shortcode wrap>before <shortcode ok /> after`, map[string]any{
		"page": testPage(),
	})

	assert.Equal(t, `<shortcode wrap>before OK after`, result)
	assert.Contains(t, logs.String(), `closing delimiter '</shortcode>' is missing`)
}

func TestRenderShortcodeSupportsSection(t *testing.T) {
	section := testSection()
	set := testShortcodeSet(map[string]*testTemplate{
		"badge": {
			execute: func(vars map[string]any) (string, error) {
				assert.Same(t, section, vars["section"])
				assert.Equal(t, "en", vars["current_lang"])
				assert.NotContains(t, vars, "page")
				return "section badge", nil
			},
		},
	})

	result := set.Render("content/_index.md", `<shortcode badge />`, map[string]any{
		"current_lang": section.Lang,
		"section":      section,
	})

	assert.Equal(t, `section badge`, result)
}

func TestHandleSectionRendersSummaryAndContent(t *testing.T) {
	section := testSection()
	section.Summary = `<shortcode badge />`
	section.Content = `<p><shortcode badge /></p>`

	h := &ShortcodeHook{
		set: testShortcodeSet(map[string]*testTemplate{
			"badge": {
				execute: func(vars map[string]any) (string, error) {
					assert.Same(t, section, vars["section"])
					return "ok", nil
				},
			},
		}),
	}

	assert.Same(t, section, h.HandleSection(section))
	assert.Equal(t, `ok`, section.Summary)
	assert.Equal(t, `<p>ok</p>`, section.Content)
}
