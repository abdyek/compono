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
	return ast.IsRuleNameOneOf(node, []string{
		"h1",
		"h2",
		"h3",
		"h4",
		"h5",
		"h6",
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
	return ast.IsRuleNameOneOf(node, []string{
		"h1-content",
		"h2-content",
		"h3-content",
		"h4-content",
		"h5-content",
		"h6-content",
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

	if tag == "p" {
		rendered := nvec.renderer.renderChildren(nvec, nvec.Node().Children())
		if ast.FindNodeByRuleName(nvec.Node().Children(), "soft-break") != nil &&
			strings.Contains(rendered, "<compono-error-block>") {
			return splitParagraphByBreakWithBlockErr(rendered)
		}

		if standaloneCompParamRefInParagraph(nvec.Node()) != nil {
			return rendered
		}
		if ast.FindNodeByRuleName(nvec.Node().Children(), "block-error") != nil {
			return renderParagraphWithBlockErrors(nvec)
		}
		return "<p>" + rendered + "</p>"
	}

	return "<" + tag + ">" + nvec.renderer.renderChildren(nvec, nvec.Node().Children()) + "</" + tag + ">"
}

func renderParagraphWithBlockErrors(nvec *nonVoidElementContent) string {
	children := nvec.Node().Children()
	if len(children) == 0 {
		return ""
	}

	result := ""
	chunk := []ast.Node{}
	hasBlockErr := ast.FindNodeByRuleName(children, "block-error") != nil

	flushChunk := func() {
		if len(chunk) == 0 {
			return
		}
		content := nvec.renderer.renderChildren(nvec, chunk)
		if content != "" {
			result += "<p>" + content + "</p>"
		}
		chunk = []ast.Node{}
	}

	for _, child := range children {
		if ast.IsRuleName(child, "block-error") {
			flushChunk()
			result += nvec.renderer.renderChildren(nvec, []ast.Node{child})
			continue
		}

		if hasBlockErr && ast.IsRuleName(child, "soft-break") {
			flushChunk()
			continue
		}

		chunk = append(chunk, child)
	}

	flushChunk()
	return result
}

func splitParagraphByBreakWithBlockErr(rendered string) string {
	parts := strings.Split(rendered, "<br>")
	result := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, "<compono-error-block>") {
			result += part
			continue
		}
		result += "<p>" + part + "</p>"
	}
	return result
}
