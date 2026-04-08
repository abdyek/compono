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
	resolved := ast.ResolveContextReferenceValue(c.renderer.root, c.Node())
	if resolved.MissingContextKey != "" {
		return `<compono-error-inline><span slot="title">Unknown key</span><span slot="description">The key <strong>` +
			resolved.MissingContextKey +
			`</strong> is not injected.</span></compono-error-inline>`
	}
	return renderResolvedValue(resolved)
}
