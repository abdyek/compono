package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type plain struct {
	baseRenderable
	renderer *renderer
}

func newPlain(rend *renderer) renderableNode {
	return &plain{
		renderer: rend,
	}
}

func (p *plain) New() renderableNode {
	return newPlain(p.renderer)
}

func (_ *plain) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "plain")
}

func (p *plain) Render() string {
	return html.EscapeString(strings.NewReplacer("\r", "", "\n", "").Replace(string(p.Node().Raw())))
}
