package html

import (
	"html"
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type err struct {
	baseRenderable
	renderer *renderer
}

func newErr(rend *renderer) renderableNode {
	return &err{
		renderer: rend,
	}
}

func (e *err) New() renderableNode {
	return newErr(e.renderer)
}

func (_ *err) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleNameOneOf(node, []string{"block-error", "inline-error"})
}

func (e *err) Render() string {
	title := ast.FindNodeByRuleName(e.Node().Children(), "error-title")
	message := ast.FindNodeByRuleName(e.Node().Children(), "error-message")

	titleStr := html.EscapeString(strings.TrimSpace(string(title.Raw())))
	messageStr := html.EscapeString(strings.TrimSpace(string(message.Raw())))

	// TODO: This is an ugly hack
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	messageStr = re.ReplaceAllString(messageStr, "<strong>$1</strong>")

	if ast.IsRuleName(e.Node(), "block-error") {
		return e.blockError(titleStr, messageStr)
	}

	return e.inlineError(titleStr, messageStr)
}

func (e *err) blockError(title, msg string) string {
	return `<compono-error-block><template shadowrootmode="closed">` +
		e.stylesheetLink() +
		`<div class="title">` +
		title +
		`</div><div class="description">` +
		msg +
		`</div></template></compono-error-block>`
}

func (e *err) inlineError(title, msg string) string {
	return `<compono-error-inline><template shadowrootmode="closed">` +
		e.stylesheetLink() +
		`<span class="title">` +
		title +
		`</span><span class="description">` +
		msg +
		`</span></template></compono-error-inline>`
}

func (e *err) stylesheetLink() string {
	if e.renderer.errorStylesheet == "" {
		return ""
	}
	return `<link rel="stylesheet" href="` + html.EscapeString(e.renderer.errorStylesheet) + `">`
}
