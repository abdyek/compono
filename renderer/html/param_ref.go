package html

// TODO: refactor for clean codes

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

func (p *paramRefInRootContent) New() renderableNode {
	return newParamRefInRootContent(p.renderer)
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
	return inlineError("Invalid parameter usage", "Parameters cannot be used in the root context.")
}

type paramRefInLocalCompDef struct {
	baseParamRef
}

func newParamRefInLocalCompDef(rend *renderer) renderableNode {
	return &paramRefInLocalCompDef{
		baseParamRef: baseParamRef{
			renderer: rend,
		},
	}
}

func (p *paramRefInLocalCompDef) New() renderableNode {
	return newParamRefInLocalCompDef(p.renderer)
}

func (_ *paramRefInLocalCompDef) Condition(invoker renderableNode, node ast.Node) bool {
	if !isRuleName(node, "param-ref") {
		return false
	}
	localCompDef := findNodeByRuleName(getAncestors(node), "local-comp-def")
	if localCompDef == nil {
		return false
	}
	return true
}

func (p *paramRefInLocalCompDef) Render() string {

	paramRefName := p.paramRefName()

	localCompDef := findNodeByRuleName(getAncestors(p.Node()), "local-comp-def")
	localCompDefHead := findNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
	compParams := findNodeByRuleName(localCompDefHead.Children(), "comp-params")

	unknownParamErr := inlineError("Unknown parameter", "The parameter <strong>"+paramRefName+"</strong> is not defined for this component.")
	if compParams == nil {
		return unknownParamErr
	}

	compParam := findNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := findNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	if compParam == nil {
		return unknownParamErr
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

type paramRefInGlobalCompDef struct {
	baseParamRef
}

func newParamRefInGlobalCompDef(rend *renderer) renderableNode {
	return &paramRefInGlobalCompDef{
		baseParamRef: baseParamRef{
			renderer: rend,
		},
	}
}

func (p *paramRefInGlobalCompDef) New() renderableNode {
	return newParamRefInGlobalCompDef(p.renderer)
}

func (_ *paramRefInGlobalCompDef) Condition(invoker renderableNode, node ast.Node) bool {
	if !isRuleName(node, "param-ref") {
		return false
	}
	localCompDef := findNodeByRuleName(getAncestors(node), "local-comp-def")
	if localCompDef != nil {
		return false
	}
	globalCompDef := findNodeByRuleName(getAncestors(node), "global-comp-def")
	if globalCompDef == nil {
		return false
	}
	return true
}

func (p *paramRefInGlobalCompDef) Render() string {

	paramRefName := p.paramRefName()

	globalCompDef := findNodeByRuleName(getAncestors(p.Node()), "global-comp-def")
	globalCompDefHead := findNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")

	unknownParamErr := inlineError("Unknown parameter", "The parameter <strong>"+paramRefName+"</strong> is not defined for this component.")

	if globalCompDefHead == nil {
		return unknownParamErr
	}

	compParams := findNodeByRuleName(globalCompDefHead.Children(), "comp-params")
	if compParams == nil {
		return unknownParamErr
	}

	compParam := findNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := findNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	if compParam == nil {
		return unknownParamErr
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
