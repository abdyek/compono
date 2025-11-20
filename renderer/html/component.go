package html

import (
	"github.com/umono-cms/compono/ast"
)

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
	compCallName := findNodeByRuleName(node.Children(), "comp-call-name")
	if compCallName == nil {
		return ""
	}

	globalCompDefAnc := findNode(getAncestors(node), func(anc ast.Node) bool {
		if !isRuleNil(anc) && anc.Rule().Name() == "global-comp-def" {
			return true
		}
		return false
	})

	localCompDefSrc := bcc.renderer.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := bcc.renderer.findLocalCompDef(localCompDefSrc, string(compCallName.Raw()))
	if localCompDef != nil {
		localCompDefContent := findNodeByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		return bcc.renderer.renderChildren(localCompDefContent.Children())
	}

	globalCompDef := bcc.renderer.findGlobalCompDef(string(compCallName.Raw()))
	if globalCompDef != nil {
		globalCompDefContent := findNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		return bcc.renderer.renderChildren(globalCompDefContent.Children())
	}

	builtinComp := bcc.renderer.findBuiltinComp(string(compCallName.Raw()))
	if builtinComp != nil {
		return builtinComp.Render(node)
	}

	return "here will be warning placeholder"
}
