package assets

import (
	"fmt"
	"strings"

	"github.com/flosch/pongo2/v7"
)

type assetNode struct {
	name    string
	pairs   map[string]pongo2.IEvaluator
	wrapper *pongo2.NodeWrapper

	hook *AssetsHook
}

func (n *assetNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) error {
	var asset *Asset

	if n.name != "" {
		a, ok := n.hook.preAssetMap[n.name]
		if !ok {
			return fmt.Errorf("assets %s is not exists", n.name)
		}
		asset = a
	} else {
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
				for name := range strings.SplitSeq(val.String(), ",") {
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
	}

	hash, err := n.hook.collectAsset(asset)
	if err != nil {
		return err
	}
	assetURL := asset.Output
	if hash != "" && asset.ShowVersion {
		assetURL = assetURL + "?" + hash[:8]
	}

	newctx := pongo2.NewChildExecutionContext(ctx)
	newctx.Private["asset_url"] = assetURL
	return n.wrapper.Execute(newctx, writer)
}

func (h *AssetsHook) assetsTagParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, error) {
	node := &assetNode{
		hook:  h,
		pairs: make(map[string]pongo2.IEvaluator),
	}

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
