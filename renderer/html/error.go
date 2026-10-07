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

	titleRawStr := strings.TrimSpace(string(title.Raw()))
	messageRawStr := strings.TrimSpace(string(message.Raw()))

	titleStr := html.EscapeString(titleRawStr)
	messageStr := html.EscapeString(messageRawStr)

	// TODO: This is an ugly hack
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	messageStr = re.ReplaceAllString(messageStr, "<strong>$1</strong>")

	marker := e.renderer.recordDiagnostic(e.Node(), titleRawStr, messageRawStr)

	if ast.IsRuleName(e.Node(), "block-error") {
		return e.blockError(titleStr, messageStr, marker)
	}

	return e.inlineError(titleStr, messageStr, marker)
}

func (e *err) blockError(title, msg, marker string) string {
	return `<compono-error-block><template shadowrootmode="closed">` +
		e.stylesheetLink() +
		`<div class="title">` +
		title +
		`</div><div class="description">` +
		msg +
		`</div></template>` + marker + `</compono-error-block>`
}

func (e *err) inlineError(title, msg, marker string) string {
	return `<compono-error-inline><template shadowrootmode="closed">` +
		e.stylesheetLink() +
		`<span class="title">` +
		title +
		`</span><span class="description">` +
		msg +
		`</span></template>` + marker + `</compono-error-inline>`
}

func (e *err) stylesheetLink() string {
	if e.renderer.errorStylesheet == "" {
		return ""
	}
	return `<link rel="stylesheet" href="` + html.EscapeString(e.renderer.errorStylesheet) + `">`
}
