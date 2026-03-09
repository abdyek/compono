package html

import (
	"regexp"
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
		rendered := normalizeRenderedMarkup(nvec.renderer.renderChildren(nvec, nvec.Node().Children()))
		if ast.FindNodeByRuleName(nvec.Node().Children(), "soft-break") != nil &&
			containsBlockLikeParagraphChild(nvec) {
			return renderParagraphWithBlockLikeChildren(nvec)
		}
		if ast.FindNodeByRuleName(nvec.Node().Children(), "soft-break") != nil &&
			strings.Contains(rendered, "<compono-error-block>") {
			return splitParagraphByBreakWithBlockErr(rendered)
		}

		if standaloneParamRef := standaloneCompParamRefInParagraph(nvec.Node()); standaloneParamRef != nil {
			if isBlockLikeRendered(rendered) || strings.HasPrefix(rendered, "<compono-error-block>") {
				return rendered
			}
		}
		if ast.FindNodeByRuleName(nvec.Node().Children(), "block-error") != nil {
			return renderParagraphWithBlockErrors(nvec)
		}
		return "<p>" + rendered + "</p>"
	}

	rendered := normalizeRenderedMarkup(nvec.renderer.renderChildren(nvec, nvec.Node().Children()))
	if strings.HasPrefix(rendered, "<compono-error-block>") {
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
			result += "<p>" + content + "</p>"
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
			if shouldOverridePreviousParagraph(rendered) {
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
		strings.HasPrefix(rendered, "<compono-error-block>")
}

func normalizeRenderedMarkup(rendered string) string {
	re := regexp.MustCompile(`\*<em>(.*?)</em>\*`)
	return re.ReplaceAllString(rendered, "<strong>$1</strong>")
}

func shouldOverridePreviousParagraph(rendered string) bool {
	if !strings.HasPrefix(rendered, "<compono-error-block>") {
		return false
	}
	return strings.Contains(rendered, "<div slot=\"title\">Invalid component usage</div>")
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
