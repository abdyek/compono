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

func (_ *plain) Condition(node ast.Node) bool {
	return isRuleName(node, "plain")
}

func (p *plain) Render(node ast.Node) string {
	return html.EscapeString(strings.TrimSpace(string(node.Raw())))
}
