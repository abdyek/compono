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

func (_ *root) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleName(node, "root")
}

func (r *root) Render() string {
	return r.renderer.renderChildren(r, r.Node().Children())
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

func (_ *rootContent) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleName(node, "root-content")
}

func (rc *rootContent) Render() string {
	return rc.renderer.renderChildren(rc, rc.Node().Children())
}
