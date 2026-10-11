package html

import "github.com/umono-cms/compono/ast"

type hr struct {
	baseRenderable
	renderer *renderer
}

func newHr(rend *renderer) renderableNode {
	return &hr{
		renderer: rend,
	}
}

func (h *hr) New() renderableNode {
	return newHr(h.renderer)
}

func (_ *hr) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "hr")
}

func (_ *hr) Render() string {
	return "<hr>"
}
