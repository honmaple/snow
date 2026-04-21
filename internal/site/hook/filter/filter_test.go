package filter

import (
	"github.com/expr-lang/expr"
	"github.com/flosch/pongo2/v7"
	"testing"
)

// BenchmarkFilterWithExpr
// BenchmarkFilterWithExpr-10		  245794		  4891 ns/op
// BenchmarkFilterWithPongo2
// BenchmarkFilterWithPongo2-10		  862621		  1277 ns/op
func BenchmarkFilterWithExpr(b *testing.B) {
	env := map[string]any{
		"tags": []string{"tag1", "tag2", "tag3"},
	}
	for b.Loop() {
		program, err := expr.Compile(`'tag2' in tags`, expr.Env(env))
		if err != nil {
			panic(err)
		}
		_, err = expr.Run(program, env)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkFilterWithPongo2(b *testing.B) {
	ctx := map[string]any{
		"tags": []string{"tag1", "tag2", "tag3"},
	}
	for b.Loop() {
		tpl, err := pongo2.FromString("{{ 'tag2' in tags }}")
		if err != nil {
			panic(err)
		}
		_, err = tpl.Execute(ctx)
		if err != nil {
			panic(err)
		}
	}
}
