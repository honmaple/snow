package snakecase

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/site/template"
)

type testTemplate struct {
	source string
}

func (t *testTemplate) Name() string {
	return "test"
}

func (t *testTemplate) Execute(ctx map[string]any) (string, error) {
	tpl, err := pongo2.FromString(t.source)
	if err != nil {
		return "", err
	}
	return tpl.Execute(ctx)
}

type testTemplateSet struct {
	tpl *testTemplate
}

func (set *testTemplateSet) Lookup(...string) template.Template {
	return set.tpl
}

func (set *testTemplateSet) FromFile(string) (template.Template, error) {
	return set.tpl, nil
}

func (set *testTemplateSet) FromBytes([]byte) (template.Template, error) {
	return set.tpl, nil
}

func (set *testTemplateSet) FromString(string) (template.Template, error) {
	return set.tpl, nil
}

func (set *testTemplateSet) Register(string, any) error {
	return nil
}

func (set *testTemplateSet) RegisterTag(string, pongo2.TagParser) error {
	return nil
}

func (set *testTemplateSet) RegisterFilter(string, pongo2.FilterFunction) error {
	return nil
}

func (set *testTemplateSet) RegisterTransient(string, template.TransientFunction) error {
	return nil
}

type testProfile struct {
	DisplayName string
	Score       int
	Nested      *testNested
	CreatedAt   time.Time
	Items       []testNested
	Labels      map[string]testNested
	hidden      string
}

func (p *testProfile) AddPoints(base int, ratio int) int {
	return base + ratio
}

func (p *testProfile) Sum(values ...int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func (p *testProfile) FirstName() string {
	return "first"
}

func (p *testProfile) LastName() string {
	return "last"
}

func (p *testProfile) WithPrefix(prefix string) (*testNested, error) {
	return &testNested{ValueName: prefix + p.DisplayName}, nil
}

func (p *testProfile) Fail() (string, error) {
	return "", errors.New("boom")
}

func (p *testProfile) Clear() error {
	return nil
}

func (p *testProfile) ArgSummary(enabled bool, count uint8, ratio float32, raw *pongo2.Value) string {
	return fmt.Sprintf("%s:%t:%d:%.1f", raw.String(), enabled, count, ratio)
}

type testNested struct {
	ValueName string
}

func (n testNested) RenderName(suffix string) string {
	return n.ValueName + suffix
}

func renderSnakeTemplate(t *testing.T, source string, ctx map[string]any) (string, error) {
	t.Helper()

	tpl := &Template{Template: &testTemplate{source: source}}
	return tpl.Execute(ctx)
}

func TestTemplateExecuteWrapsFieldsAndMethodsAsSnakeCase(t *testing.T) {
	createdAt := time.Date(2026, 6, 10, 11, 12, 13, 0, time.UTC)
	result, err := renderSnakeTemplate(t, `{{ profile.display_name }}|{{ profile.score }}|{{ profile.hidden }}|{{ profile.nested.value_name }}|{{ profile.add_points(2, 3) }}|{{ profile.sum(1, 2, 3) }}|{{ profile.first_name() }}|{{ profile.last_name() }}|{{ profile.with_prefix("hi ").value_name }}|{{ profile.created_at|date:"2006-01-02" }}`, map[string]any{
		"profile": &testProfile{
			DisplayName: "Ada",
			Score:       42,
			Nested:      &testNested{ValueName: "inner"},
			CreatedAt:   createdAt,
			hidden:      "private",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "Ada|42||inner|5|6|first|last|hi Ada|2026-06-10"
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
}

func TestTemplateExecuteWrapsSlicesAndStringKeyedMaps(t *testing.T) {
	result, err := renderSnakeTemplate(t, `{{ profile.items.0.value_name }}|{{ profile.items.0.render_name("!") }}|{{ profile.labels.main.value_name }}|{{ extra.child.value_name }}`, map[string]any{
		"profile": testProfile{
			Items: []testNested{
				{ValueName: "slice"},
			},
			Labels: map[string]testNested{
				"main": {ValueName: "map"},
			},
		},
		"extra": map[string]testNested{
			"child": {ValueName: "extra"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "slice|slice!|map|extra"
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
}

func TestTemplateExecuteReturnsMethodErrors(t *testing.T) {
	_, err := renderSnakeTemplate(t, `{{ profile.fail() }}`, map[string]any{
		"profile": &testProfile{},
	})
	if err == nil {
		t.Fatal("expected method error")
	}
}

func TestTemplateExecuteHandlesNilPointerFields(t *testing.T) {
	result, err := renderSnakeTemplate(t, `{{ profile.nested.value_name }}|{{ profile.clear() }}`, map[string]any{
		"profile": &testProfile{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result != "|" {
		t.Fatalf("result = %q, want %q", result, "|")
	}
}

func TestTemplateExecuteWrapsFunctionReturnValues(t *testing.T) {
	result, err := renderSnakeTemplate(t, `{{ get_profile("Ada").display_name }}|{{ get_profile("Ada").with_prefix("hi ").value_name }}|{{ list_profiles().0.display_name }}|{{ get_nested().value_name }}|{{ get_labels().main.value_name }}|{{ get_map().child.value_name }}`, map[string]any{
		"get_profile": func(name string) *testProfile {
			return &testProfile{DisplayName: name}
		},
		"list_profiles": func() []testProfile {
			return []testProfile{{DisplayName: "Grace"}}
		},
		"get_nested": func() (*testNested, error) {
			return &testNested{ValueName: "nested"}, nil
		},
		"get_labels": func() map[string]testNested {
			return map[string]testNested{"main": {ValueName: "label"}}
		},
		"get_map": func() map[string]any {
			return map[string]any{"child": &testNested{ValueName: "map"}}
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "Ada|hi Ada|Grace|nested|label|map"
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
}

func TestTemplateExecuteWrapsFunctionReturnTimeWithoutSnakeMapping(t *testing.T) {
	result, err := renderSnakeTemplate(t, `{{ get_time()|date:"2006-01-02" }}`, map[string]any{
		"get_time": func() time.Time {
			return time.Date(2026, 6, 10, 11, 12, 13, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result != "2026-06-10" {
		t.Fatalf("result = %q, want %q", result, "2026-06-10")
	}
}

func TestTemplateExecuteConvertsFunctionAndMethodArguments(t *testing.T) {
	result, err := renderSnakeTemplate(t, `{{ summarize(true, 7, 2.5, "raw") }}|{{ profile.arg_summary(true, 8, 1.5, "value") }}`, map[string]any{
		"profile": &testProfile{},
		"summarize": func(enabled bool, count uint8, ratio float32, raw *pongo2.Value) string {
			return fmt.Sprintf("%s:%t:%d:%.1f", raw.String(), enabled, count, ratio)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := "raw:true:7:2.5|value:true:8:1.5"
	if result != want {
		t.Fatalf("result = %q, want %q", result, want)
	}
}

func TestTemplateExecuteSupportsNilFunctionErrorReturn(t *testing.T) {
	result, err := renderSnakeTemplate(t, `{{ noop() }}ok`, map[string]any{
		"noop": func() error {
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result != "ok" {
		t.Fatalf("result = %q, want %q", result, "ok")
	}
}

func TestTemplateExecuteReturnsFunctionErrors(t *testing.T) {
	_, err := renderSnakeTemplate(t, `{{ get_profile().display_name }}`, map[string]any{
		"get_profile": func() (*testProfile, error) {
			return nil, errors.New("not found")
		},
	})
	if err == nil {
		t.Fatal("expected function error")
	}
}

func TestTemplateExecuteReturnsArgumentAndSignatureErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
		ctx    map[string]any
	}{
		{
			name:   "too few arguments",
			source: `{{ add(1) }}`,
			ctx: map[string]any{
				"add": func(a int, b int) int { return a + b },
			},
		},
		{
			name:   "wrong argument type",
			source: `{{ needs_struct("nope") }}`,
			ctx: map[string]any{
				"needs_struct": func(n testNested) string { return n.ValueName },
			},
		},
		{
			name:   "unsupported return signature",
			source: `{{ invalid_return() }}`,
			ctx: map[string]any{
				"invalid_return": func() (string, int) { return "value", 1 },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := renderSnakeTemplate(t, tt.source, tt.ctx)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestTemplateSetWrapsTemplatesFromEveryFactoryMethod(t *testing.T) {
	set := &TemplateSet{TemplateSet: &testTemplateSet{
		tpl: &testTemplate{source: `{{ profile.display_name }}`},
	}}
	ctx := map[string]any{
		"profile": &testProfile{DisplayName: "Ada"},
	}

	tests := []struct {
		name string
		tpl  template.Template
		err  error
	}{
		{name: "lookup", tpl: set.Lookup("test")},
	}
	fromFile, err := set.FromFile("test")
	tests = append(tests, struct {
		name string
		tpl  template.Template
		err  error
	}{name: "from file", tpl: fromFile, err: err})

	fromBytes, err := set.FromBytes([]byte("test"))
	tests = append(tests, struct {
		name string
		tpl  template.Template
		err  error
	}{name: "from bytes", tpl: fromBytes, err: err})

	fromString, err := set.FromString("test")
	tests = append(tests, struct {
		name string
		tpl  template.Template
		err  error
	}{name: "from string", tpl: fromString, err: err})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				t.Fatal(tt.err)
			}
			if tt.tpl == nil {
				t.Fatal("expected template")
			}
			result, err := tt.tpl.Execute(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if result != "Ada" {
				t.Fatalf("result = %q, want %q", result, "Ada")
			}
		})
	}
}
