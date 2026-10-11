package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
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
		return renderParagraph(nvec)
	}

	rendered := nvec.renderer.renderChildren(nvec, nvec.Node().Children())

	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		rendered = util.TrimWhitespace(rendered)
	}

	if isEmptyRendered(rendered) {
		return rendered
	}
	if hasBlockErrorPrefix(rendered) {
		return rendered
	}
	return "<" + tag + ">" + rendered + "</" + tag + ">"
}

// renderParagraph renders a paragraph. A block line splits it: the paragraph lines before it and after it form separate paragraphs and the block line is rendered between them, outside any <p>.
func renderParagraph(nvec *nonVoidElementContent) string {
	lines := [][]ast.Node{}
	current := []ast.Node{}
	for _, child := range nvec.Node().Children() {
		if ast.IsRuleName(child, "soft-break") {
			lines = append(lines, current)
			current = []ast.Node{}
			continue
		}
		current = append(current, child)
	}
	lines = append(lines, current)

	result := ""
	rendered := []string{}
	flush := func() {
		if len(rendered) > 0 {
			result += paragraph(util.TrimWhitespace(strings.Join(rendered, "<br>")))
			rendered = nil
		}
	}

	for _, line := range lines {
		nonEmpty := []ast.Node{}
		for _, node := range line {
			if ast.IsRuleName(node, "plain") && util.TrimWhitespace(string(node.Raw())) == "" {
				continue
			}
			nonEmpty = append(nonEmpty, node)
		}

		isBlockLine := false
		var thatNode ast.Node
		if len(nonEmpty) == 1 {
			node := nonEmpty[0]
			if ast.IsRuleName(node, "block-error") || isCompParamRef(nvec.renderer, nvec, node) {
				isBlockLine = true
				thatNode = node
			}
		}

		if isBlockLine {
			flush()
			result += nvec.renderer.renderChildren(nvec, []ast.Node{thatNode})
		} else {
			rendered = append(rendered, util.TrimWhitespace(nvec.renderer.renderChildren(nvec, line)))
		}
	}
	flush()
	return result
}

// paragraph wraps the rendered inline content of a paragraph in <p>. A line
// that resolves empty is dropped with its line break, and a paragraph whose
// lines are all empty renders only its markers.
func paragraph(content string) string {
	parts := strings.Split(content, "<br>")
	out, pending := "", ""
	wrote := false
	for _, part := range parts {
		if isEmptyRendered(part) {
			pending += strings.NewReplacer(" ", "", "\t", "", "\n", "").Replace(part)
			continue
		}
		if wrote {
			out += "<br>"
		}
		out += pending + part
		pending = ""
		wrote = true
	}
	out += pending
	if !wrote {
		return out
	}
	return "<p>" + out + "</p>"
}
