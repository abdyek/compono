package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

type nonVoidElement struct {
	baseRenderable
	renderer *renderer
}

func newNonVoidElement(rend *renderer) renderableNode {
	return &nonVoidElement{
		renderer: rend,
	}
}

func (_ *nonVoidElement) Condition(node ast.Node) bool {
	return isRuleNameOneOf(node, []string{
		"h1",
		"h2",
		"p",
		"em",
		"strong",
	})
}

func (nve *nonVoidElement) Render(node ast.Node) string {
	return nve.renderer.renderChildren(nve, node.Children())
}

type nonVoidElementContent struct {
	baseRenderable
	renderer *renderer
}

func newNonVoidElementContent(rend *renderer) renderableNode {
	return &nonVoidElementContent{
		renderer: rend,
	}
}

func (_ *nonVoidElementContent) Condition(node ast.Node) bool {
	return isRuleNameOneOf(node, []string{
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
	return "<" + tag + ">" + nvec.renderer.renderChildren(nvec, node.Children()) + "</" + tag + ">"
}
