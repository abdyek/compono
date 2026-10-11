package html

import (
	"github.com/umono-cms/compono/ast"
)

type blockquote struct {
	baseRenderable
	renderer *renderer
}

func newBlockquote(rend *renderer) renderableNode {
	return &blockquote{
		renderer: rend,
	}
}

func (b *blockquote) New() renderableNode {
	return newBlockquote(b.renderer)
}

func (_ *blockquote) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "blockquote")
}

func (b *blockquote) Render() string {
	content := b.renderer.renderChildren(b, b.Node().Children())
	if isEmptyRendered(content) {
		return content
	}
	return "<blockquote>" + content + "</blockquote>"
}
