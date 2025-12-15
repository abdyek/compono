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

func (nve *nonVoidElement) New() renderableNode {
	return newNonVoidElement(nve.renderer)
}

func (_ *nonVoidElement) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleNameOneOf(node, []string{
		"h1",
		"h2",
		"p",
		"em",
		"strong",
	})
}

func (nve *nonVoidElement) Render() string {
	return nve.renderer.renderChildren(nve, nve.Node().Children())
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

func (nvec *nonVoidElementContent) New() renderableNode {
	return newNonVoidElementContent(nvec.renderer)
}

func (_ *nonVoidElementContent) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleNameOneOf(node, []string{
		"h1-content",
		"h2-content",
		"p-content",
		"em-content",
		"strong-content",
	})
}

func (nvec *nonVoidElementContent) Render() string {
	rule := nvec.Node().Rule()

	if rule == nil {
		return ""
	}

	name := rule.Name()
	idx := strings.Index(name, "-")

	if idx == -1 {
		return ""
	}

	tag := name[:idx]
	return "<" + tag + ">" + nvec.renderer.renderChildren(nvec, nvec.Node().Children()) + "</" + tag + ">"
}
