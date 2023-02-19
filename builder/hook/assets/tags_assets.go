package assets

import (
	"strings"

	"github.com/flosch/pongo2/v6"
)

type assetNode struct {
	name    string
	pairs   map[string]pongo2.IEvaluator
	wrapper *pongo2.NodeWrapper
	assets  *assets
}

func (node *assetNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	assetURL := ""
	if node.name == "" {
		opt := option{}
		for key, value := range node.pairs {
			val, err := value.Evaluate(ctx)
			if err != nil {
				return err
			}
			switch key {
			case "files":
				opt.files = strings.Split(val.String(), ",")
			case "filters":
				opt.filters = strings.Split(val.String(), ",")
			case "output":
				opt.output = val.String()
			}
		}
		opt.filterOpts = make([]filterOption, len(opt.filters))
		if err := node.assets.execute(opt); err != nil {
			return &pongo2.Error{Sender: "tag:assets", OrigError: err}
		}
		assetURL = opt.output
	} else {
		assetURL = node.assets.conf.GetString("params.assets." + node.name + ".output")
	}
	newctx := pongo2.NewChildExecutionContext(ctx)
	newctx.Private["asset_url"] = assetURL
	return node.wrapper.Execute(newctx, writer)
}

func (ws *assets) assetParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	node := &assetNode{pairs: make(map[string]pongo2.IEvaluator)}

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
