package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type plain struct {
	renderer *renderer
}

func newPlain(rend *renderer) renderableNode {
	return &plain{
		renderer: rend,
	}
}

func (p *plain) Condition(node ast.Node) bool {
	return p.renderer.isRuleName(node, "plain")
}

func (p *plain) Render(node ast.Node) string {
	return html.EscapeString(strings.TrimSpace(string(node.Raw())))
}
