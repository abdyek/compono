package html

import (
	"strconv"
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
	rendered := nvec.renderer.renderChildren(nvec, nvec.Node().Children())

	switch tag {
	case "p", "h1", "h2", "h3", "h4", "h5", "h6":
		rendered = util.TrimWhitespace(rendered)
	}

	if tag == "p" {
		if ast.FindNodeByRuleName(nvec.Node().Children(), "soft-break") != nil &&
			containsBlockLikeParagraphChild(nvec) {
			return renderParagraphWithBlockLikeChildren(nvec)
		}
		if ast.FindNodeByRuleName(nvec.Node().Children(), "soft-break") != nil &&
			containsBlockError(rendered) {
			return splitParagraphByBreakWithBlockErr(rendered)
		}

		if standaloneParamRef := standaloneCompParamRefInParagraph(nvec.Node()); standaloneParamRef != nil {
			if isBlockLikeRendered(rendered) || hasBlockErrorPrefix(rendered) {
				return rendered
			}
		}
		if ast.FindNodeByRuleName(nvec.Node().Children(), "block-error") != nil {
			return renderParagraphWithBlockErrors(nvec)
		}
		return paragraph(rendered)
	}

	if isEmptyRendered(rendered) {
		return rendered
	}
	if hasBlockErrorPrefix(rendered) {
		return rendered
	}
	return "<" + tag + ">" + rendered + "</" + tag + ">"
}

func containsBlockLikeParagraphChild(nvec *nonVoidElementContent) bool {
	for _, child := range nvec.Node().Children() {
		if ast.IsRuleName(child, "soft-break") {
			continue
		}
		rendered := nvec.renderer.renderChildren(nvec, []ast.Node{child})
		if isBlockLikeRendered(rendered) {
			return true
		}
	}
	return false
}

func renderParagraphWithBlockLikeChildren(nvec *nonVoidElementContent) string {
	result := ""
	inlineChunk := []ast.Node{}

	flush := func() {
		if len(inlineChunk) == 0 {
			return
		}
		content := nvec.renderer.renderChildren(nvec, inlineChunk)
		if content != "" {
			result += paragraph(content)
		}
		inlineChunk = []ast.Node{}
	}

	for _, child := range nvec.Node().Children() {
		if ast.IsRuleName(child, "soft-break") {
			flush()
			continue
		}

		rendered := nvec.renderer.renderChildren(nvec, []ast.Node{child})
		if isBlockLikeRendered(rendered) {
			flush()
			if nvec.renderer.shouldOverridePreviousParagraph(rendered) {
				result = rendered
				inlineChunk = []ast.Node{}
				continue
			}
			result += rendered
			continue
		}

		inlineChunk = append(inlineChunk, child)
	}

	flush()
	return result
}

func isBlockLikeRendered(rendered string) bool {
	return strings.HasPrefix(rendered, "<h1>") ||
		strings.HasPrefix(rendered, "<h2>") ||
		strings.HasPrefix(rendered, "<h3>") ||
		strings.HasPrefix(rendered, "<h4>") ||
		strings.HasPrefix(rendered, "<h5>") ||
		strings.HasPrefix(rendered, "<h6>") ||
		strings.HasPrefix(rendered, "<p>") ||
		strings.HasPrefix(rendered, "<img ") ||
		strings.HasPrefix(rendered, "<picture>") ||
		strings.HasPrefix(rendered, "<header") ||
		strings.HasPrefix(rendered, "<main") ||
		strings.HasPrefix(rendered, "<footer") ||
		strings.HasPrefix(rendered, "<article") ||
		strings.HasPrefix(rendered, "<aside") ||
		strings.HasPrefix(rendered, "<nav") ||
		strings.HasPrefix(rendered, "<details") ||
		strings.HasPrefix(rendered, "<ul") ||
		strings.HasPrefix(rendered, "<ol") ||
		strings.HasPrefix(rendered, "<hr") ||
		strings.HasPrefix(rendered, "<blockquote") ||
		hasBlockErrorPrefix(rendered)
}

func (r *renderer) shouldOverridePreviousParagraph(rendered string) bool {
	if !hasBlockErrorPrefix(rendered) {
		return false
	}
	for _, m := range diagnosticMarkerPattern.FindAllStringSubmatch(rendered, -1) {
		if !strings.HasPrefix(m[0], blockDiagnosticMarkerPrefix) {
			continue
		}
		if len(m) < 2 {
			continue
		}
		idx, err := strconv.Atoi(m[1])
		if err != nil || idx < 0 || idx >= len(r.diagnostics) {
			continue
		}
		if r.diagnostics[idx].Title == "Invalid component usage" {
			return true
		}
	}
	return false
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
			result += paragraph(content)
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

func splitParagraphByBreakWithBlockErr(rendered string) string {
	parts := strings.Split(rendered, "<br>")
	result := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if hasBlockErrorPrefix(part) {
			result += part
			continue
		}
		result += paragraph(part)
	}
	return result
}
