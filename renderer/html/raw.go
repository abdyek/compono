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
	if !isRuleName(invoker.Node(), "code-block-content") {
		return false
	}
	return isRuleName(node, "plain")
}

func (r *raw) Render() string {
	return html.EscapeString(string(r.Node().Raw()))
}
