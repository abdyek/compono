package html

import "github.com/umono-cms/compono/ast"

// TODO: complete it
type blockCompCall struct {
	renderer *renderer
}

func newBlockCompCall(rend *renderer) renderableNode {
	return &blockCompCall{
		renderer: rend,
	}
}

func (bcc *blockCompCall) Condition(node ast.Node) bool {
	return false
}

func (bcc *blockCompCall) Render(node ast.Node) string {
	return ""
}
