package html

import (
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

	if isRuleName(e.Node(), "block-error") {
		return blockError(titleStr, messageStr)
	}

	return inlineError(titleStr, messageStr)
}
