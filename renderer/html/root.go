package html

import (
	"github.com/umono-cms/compono/ast"
)

type root struct {
	baseRenderable
	renderer *renderer
}

func newRoot(rend *renderer) renderableNode {
	return &root{
		renderer: rend,
	}
}

func (_ *root) Condition(node ast.Node) bool {
	return isRuleName(node, "root")
}

func (r *root) Render(node ast.Node) string {
	return r.renderer.renderChildren(r, node.Children())
}

type rootContent struct {
	baseRenderable
	renderer *renderer
}

func newRootContent(rend *renderer) renderableNode {
	return &rootContent{
		renderer: rend,
	}
}

func (_ *rootContent) Condition(node ast.Node) bool {
	return isRuleName(node, "root-content")
}

func (rc *rootContent) Render(node ast.Node) string {
	return rc.renderer.renderChildren(rc, node.Children())
}
