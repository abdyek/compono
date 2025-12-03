package html

// TODO: refactor for clean coding

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

type baseParamRef struct {
	baseRenderable
	renderer *renderer
}

func (bpr *baseParamRef) paramRefName() string {
	paramRefName := findNodeByRuleName(bpr.Node().Children(), "param-ref-name")
	return strings.TrimSpace(string(paramRefName.Raw()))
}

type paramRefInRootContent struct {
	baseRenderable
	renderer *renderer
}

func newParamRefInRootContent(rend *renderer) renderableNode {
	return &paramRefInRootContent{
		renderer: rend,
	}
}

func (_ *paramRefInRootContent) Condition(invoker renderableNode, node ast.Node) bool {
	if !isRuleName(node, "param-ref") {
		return false
	}
	compDefContent := findNode(getAncestors(node), func(anc ast.Node) bool {
		return isRuleNameOneOf(anc, []string{"local-comp-def-content", "global-comp-def-content"})
	})
	if compDefContent != nil {
		return false
	}
	return true
}

func (_ *paramRefInRootContent) Render() string {
	return "error placeholder"
}

type paramRefInLocalCompDefOfRoot struct {
	baseParamRef
}

func newParamRefInLocalCompDefOfRoot(rend *renderer) renderableNode {
	return &paramRefInLocalCompDefOfRoot{
		baseParamRef: baseParamRef{
			renderer: rend,
		},
	}
}

func (_ *paramRefInLocalCompDefOfRoot) Condition(invoker renderableNode, node ast.Node) bool {
	if !isRuleName(node, "param-ref") {
		return false
	}
	localCompDef := findNodeByRuleName(getAncestors(node), "local-comp-def")
	if localCompDef == nil {
		return false
	}
	globalCompDef := findNodeByRuleName(getAncestors(node), "global-comp-def")
	if globalCompDef != nil {
		return false
	}
	return true
}

func (p *paramRefInLocalCompDefOfRoot) Render() string {

	paramRefName := p.paramRefName()

	localCompDef := findNodeByRuleName(getAncestors(p.Node()), "local-comp-def")
	localCompDefHead := findNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
	compParams := findNodeByRuleName(localCompDefHead.Children(), "comp-params")

	if compParams == nil {
		return "invalid! the params not found"
	}

	compParam := findNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := findNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	if compParam == nil {
		return "invalid! the param not found"
	}

	compCall := findNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
		return isRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	compCallArgs := findNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgs != nil {
		compCallArg := findNode(compCallArgs.Children(), func(cca ast.Node) bool {
			argName := findNodeByRuleName(cca.Children(), "comp-call-arg-name")
			if strings.TrimSpace(string(argName.Raw())) == paramRefName {
				return true
			}
			return false
		})
		if compCallArg != nil {
			argValue := findNodeByRuleName(findNode(findNodeByRuleName(compCallArg.Children(), "comp-call-arg-type").Children(), func(node ast.Node) bool {
				return isRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg"})
			}).Children(), "comp-call-arg-value")
			return strings.TrimSpace(string(argValue.Raw()))
		}
	}

	compParamDefaValue := findNodeByRuleName(findNode(findNodeByRuleName(compParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return isRuleNameOneOf(node, []string{"comp-string-param", "comp-call-param", "comp-bool-param"})
	}).Children(), "comp-param-defa-value")

	return strings.TrimSpace(string(compParamDefaValue.Raw()))
}
