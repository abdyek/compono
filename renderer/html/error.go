package html

import (
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
	return isRuleNameOneOf(node, []string{"block-error", "inline-error"})
}

func (e *err) Render() string {
	title := findNodeByRuleName(e.Node().Children(), "error-title")
	message := findNodeByRuleName(e.Node().Children(), "error-message")

	titleStr := strings.TrimSpace(string(title.Raw()))
	messageStr := strings.TrimSpace(string(message.Raw()))

	// TODO: This is an ugly hack
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	messageStr = re.ReplaceAllString(messageStr, "<strong>$1</strong>")

	if isRuleName(e.Node(), "block-error") {
		return e.blockError(titleStr, messageStr)
	}

	return e.inlineError(titleStr, messageStr)
}

func (e *err) blockError(title, msg string) string {
	return `<compono-error-block><div slot="title">` +
		title +
		`</div><div slot="description">` +
		msg +
		`</div></compono-error-block>`
}

func (e *err) inlineError(title, msg string) string {
	return `<compono-error-inline><span slot="title">` +
		title +
		`</span><span slot="description">` +
		msg +
		`</span></compono-error-inline>`
}
