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
	return ast.IsRuleNameOneOf(node, []string{"block-error", "inline-error"})
}

func (e *err) Render() string {
	title := ast.FindNodeByRuleName(e.Node().Children(), "error-title")
	message := ast.FindNodeByRuleName(e.Node().Children(), "error-message")

	titleRawStr := strings.TrimSpace(string(title.Raw()))
	messageRawStr := strings.TrimSpace(string(message.Raw()))

	return e.renderer.recordDiagnostic(e.Node(), titleRawStr, messageRawStr, ast.IsRuleName(e.Node(), "block-error"))
}
