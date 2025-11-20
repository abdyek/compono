package html

import (
	"github.com/umono-cms/compono/ast"
)

// TODO: complete it
type blockCompCall struct {
	renderer *renderer
}

func newBlockCompCall(rend *renderer) renderableNode {
	return &blockCompCall{
		renderer: rend,
	}
}

func (_ *blockCompCall) Condition(node ast.Node) bool {
	return isRuleName(node, "block-comp-call")
}

func (bcc *blockCompCall) Render(node ast.Node) string {
	compCallName := findChildByRuleName(node.Children(), "comp-call-name")
	if compCallName == nil {
		return ""
	}

	// TODO: Add a condition for localCompDef(s) of globalCompDef(s)

	localCompDef := bcc.renderer.findLocalCompDef(string(compCallName.Raw()))
	if localCompDef != nil {
		localCompDefContent := findChildByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		return bcc.renderer.renderChildren(localCompDefContent.Children())
	}

	globalCompDef := bcc.renderer.findGlobalCompDef(string(compCallName.Raw()))
	if globalCompDef != nil {
		globalCompDefContent := findChildByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		return bcc.renderer.renderChildren(globalCompDefContent.Children())
	}

	// TODO: Add built-in component solution here

	return "here will be warning placeholder"
}
