package html

import "github.com/umono-cms/compono/ast"

type contextRef struct {
	baseRenderable
	renderer *renderer
}

func newContextRef(rend *renderer) renderableNode {
	return &contextRef{renderer: rend}
}

func (c *contextRef) New() renderableNode {
	return newContextRef(c.renderer)
}

func (_ *contextRef) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "context-ref")
}

func (c *contextRef) Render() string {
	return renderResolvedValue(ast.ResolveContextReferenceValue(c.renderer.root, c.Node()))
}
