package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

type nonVoidElement struct {
	renderer *renderer
}

func newNonVoidElement(rend *renderer) renderableNode {
	return &nonVoidElement{
		renderer: rend,
	}
}

func (nve *nonVoidElement) Condition(node ast.Node) bool {
	return nve.renderer.inRuleName(node, []string{
		"h1",
		"h2",
		"p",
		"em",
		"strong",
	})
}

func (nve *nonVoidElement) Render(node ast.Node) string {
	return nve.renderer.renderChildren(node.Children())
}

type nonVoidElementContent struct {
	renderer *renderer
}

func newNonVoidElementContent(rend *renderer) renderableNode {
	return &nonVoidElementContent{
		renderer: rend,
	}
}

func (nvec *nonVoidElementContent) Condition(node ast.Node) bool {
	return nvec.renderer.inRuleName(node, []string{
		"h1-content",
		"h2-content",
		"p-content",
		"em-content",
		"strong-content",
	})
}

func (nvec *nonVoidElementContent) Render(node ast.Node) string {
	rule := node.Rule()

	if rule == nil {
		return ""
	}

	name := rule.Name()
	idx := strings.Index(name, "-")

	if idx == -1 {
		return ""
	}

	tag := name[:idx]
	return "<" + tag + ">" + nvec.renderer.renderChildren(node.Children()) + "</" + tag + ">"
}
