package html

import "github.com/umono-cms/compono/ast"

type br struct {
	baseRenderable
	renderer *renderer
}

func newBr(rend *renderer) renderableNode {
	return &br{
		renderer: rend,
	}
}

func (b *br) New() renderableNode {
	return newBr(b.renderer)
}

func (_ *br) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "soft-break")
}

func (_ *br) Render() string {
	return "<br>"
}
