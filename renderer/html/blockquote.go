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
	content := ""
	for _, child := range b.Node().Children() {
		rendered := b.renderer.renderChildren(b, []ast.Node{child})
		if len(rendered) >= 7 && rendered[:3] == "<p>" && rendered[len(rendered)-4:] == "</p>" {
			inner := rendered[3 : len(rendered)-4]
			if isEmptyRendered(inner) {
				content += inner
				continue
			}
		}
		content += rendered
	}
	if isEmptyRendered(content) {
		return content
	}
	return "<blockquote>" + content + "</blockquote>"
}
