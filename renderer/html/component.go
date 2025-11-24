package html

import (
	"github.com/umono-cms/compono/ast"
)

type compCall struct {
	renderer *renderer
}

func newCompCall(rend *renderer) renderableNode {
	return &compCall{
		renderer: rend,
	}
}

func (_ *compCall) Condition(node ast.Node) bool {
	return isRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
}

func (cc *compCall) Render(node ast.Node) string {
	inlineCompCall := false
	if isRuleName(node, "inline-comp-call") {
		inlineCompCall = true
	}

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

	localCompDefSrc := cc.renderer.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := cc.renderer.findLocalCompDef(localCompDefSrc, string(compCallName.Raw()))
	if localCompDef != nil {
		localCompDefContent := findNodeByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		if inlineCompCall {
			return cc.renderInlineCompCall(localCompDefContent)
		}
		return cc.renderer.renderChildren(localCompDefContent.Children())
	}

	globalCompDef := cc.renderer.findGlobalCompDef(string(compCallName.Raw()))
	if globalCompDef != nil {
		globalCompDefContent := findNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		if inlineCompCall {
			return cc.renderInlineCompCall(globalCompDefContent)
		}
		return cc.renderer.renderChildren(globalCompDefContent.Children())
	}

	builtinComp := cc.renderer.findBuiltinComp(string(compCallName.Raw()))
	if builtinComp != nil {
		return builtinComp.Render(node)
	}

	return "here will be warning placeholder"
}

func (cc *compCall) renderInlineCompCall(compDefContent ast.Node) string {
	childCount := len(compDefContent.Children())
	if childCount == 0 {
		return ""
	}
	if childCount > 1 || findNodeByRuleName(compDefContent.Children(), "p") == nil {
		return "Block components are disallowed inside inline components."
	}
	return cc.renderer.renderChildren(compDefContent.Children())
}
