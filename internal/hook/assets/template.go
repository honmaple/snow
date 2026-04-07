package assets

import (
	"errors"
	"reflect"
	"strings"

	"github.com/flosch/pongo2/v7"
	"github.com/honmaple/snow/internal/theme/template"
	"github.com/honmaple/snow/internal/core"
	"github.com/spf13/cast"
)

type assetNode struct {
	ctx     *core.Context
	name    string
	pairs   map[string]pongo2.IEvaluator
	wrapper *pongo2.NodeWrapper
}

func (n *assetNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) error {
	collector, ok := ctx.Public["__assets_collector"].(*AssetsCollector)
	if !ok {
		return errors.New("can't found assets collector")
	}

	asset := &Asset{}
	if n.name == "" {
		for key, value := range n.pairs {
			val, err := value.Evaluate(ctx)
			if err != nil {
				return err
			}
			switch key {
			case "files":
				asset.Files = strings.Split(val.String(), ",")
			case "filters":
				asset.Filters = make([]map[string]map[string]any, 0)
				for _, name := range strings.Split(val.String(), ",") {
					asset.Filters = append(asset.Filters, map[string]map[string]any{
						name: nil,
					})
				}
			case "output":
				asset.Output = val.String()
			case "version":
				asset.ShowVersion = val.Bool()
			}
		}
	} else {
		conf := n.ctx.Config.Sub("hooks.assets." + n.name)

		asset.Files = conf.GetStringSlice("files")
		asset.Output = conf.GetString("output")
		asset.ShowVersion = conf.GetBool("version")
		if m := conf.Get("filters"); m != nil {
			switch reflect.TypeOf(m).Kind() {
			case reflect.Slice:
				// - libsass:
				//     path: ""
				// - cssmin:
				asset.Filters = make([]map[string]map[string]any, 0)
				for _, item := range m.([]any) {
					for k, v := range cast.ToStringMap(item) {
						asset.Filters = append(asset.Filters, map[string]map[string]any{
							k: cast.ToStringMap(v),
						})
						break
					}
				}
			case reflect.String:
				// libcass,css
				asset.Filters = make([]map[string]map[string]any, 0)
				for _, name := range strings.Split(m.(string), ",") {
					asset.Filters = append(asset.Filters, map[string]map[string]any{
						name: nil,
					})
				}
			}
		}
	}
	hash, err := asset.GetHash(n.ctx)
	if err != nil {
		return err
	}
	if hash != "" && asset.ShowVersion {
		asset.Output = asset.Output + "?" + hash[:8]
	}
	if err := collector.Collect(asset); err != nil {
		return err
	}

	newctx := pongo2.NewChildExecutionContext(ctx)
	newctx.Private["asset_url"] = asset
	return n.wrapper.Execute(newctx, writer)
}

func assetsTagParser(ctx *core.Context) pongo2.TagParser {
	return func(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, error) {
		node := &assetNode{ctx: ctx, pairs: make(map[string]pongo2.IEvaluator)}

		wrapper, endargs, err := doc.WrapUntilTag("endassets")
		if err != nil {
			return nil, err
		}
		node.wrapper = wrapper

		if endargs.Count() > 0 {
			return nil, endargs.Error("Arguments not allowed here.", nil)
		}

		// {% assets %}
		// {% endassets %}
		if arguments.Count() == 0 {
			return nil, arguments.Error("Tag 'assets' requires at least one argument.", nil)
		}

		// {% assets css %}
		//   <link rel="stylesheet" href="{{ config.site.url }}/{{ asset_url }}">
		// {% endassets %}
		if token := arguments.MatchType(pongo2.TokenString); token != nil {
			node.name = token.Val
			return node, nil
		}

		// {% assets files="css/style.scss" filters="libsass,cssmin" output="css/style.min.css" %}
		//   <link rel="stylesheet" href="{{ config.site.url }}/{{ asset_url }}">
		// {% endassets %}
		for arguments.Remaining() > 0 {
			keyToken := arguments.MatchType(pongo2.TokenIdentifier)
			if keyToken == nil {
				return nil, arguments.Error("Expected an identifier", nil)
			}
			if arguments.Match(pongo2.TokenSymbol, "=") == nil {
				return nil, arguments.Error("Expected '='.", nil)
			}
			valueExpr, err := arguments.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.pairs[keyToken.Val] = valueExpr
		}
		return node, nil
	}
}

func init() {
	template.RegisterContextTag("assets", assetsTagParser)
}
