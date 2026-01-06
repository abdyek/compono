package html

import (
	"html"

	"github.com/umono-cms/compono/ast"
)

type raw struct {
	baseRenderable
	renderer *renderer
}

func newRaw(rend *renderer) renderableNode {
	return &raw{
		renderer: rend,
	}
}

func (r *raw) New() renderableNode {
	return newRaw(r.renderer)
}

func (_ *raw) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleName(node, "raw")
}

func (r *raw) Render() string {
	return html.EscapeString(string(r.Node().Raw()))
}
