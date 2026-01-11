package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

type baseParamRef struct {
	baseRenderable
	renderer *renderer
}

func (bpr *baseParamRef) paramRefName() string {
	paramRefName := ast.FindNodeByRuleName(bpr.Node().Children(), "param-ref-name")
	return strings.TrimSpace(string(paramRefName.Raw()))
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
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}
	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(node), "local-comp-def")
	if localCompDef == nil {
		return false
	}
	return true
}

func (p *paramRefInLocalCompDef) Render() string {

	paramRefName := p.paramRefName()

	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "local-comp-def")
	localCompDefHead := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
	compParams := ast.FindNodeByRuleName(localCompDefHead.Children(), "comp-params")

	compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	compCall := ast.FindNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgs != nil {
		compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
			argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
			if strings.TrimSpace(string(argName.Raw())) == paramRefName {
				return true
			}
			return false
		})
		if compCallArg != nil {
			argValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type").Children(), func(node ast.Node) bool {
				return ast.IsRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg"})
			}).Children(), "comp-call-arg-value")
			if argValue == nil {
				return ""
			}
			return strings.TrimSpace(string(argValue.Raw()))
		}
	}

	compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(compParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-call-param", "comp-bool-param"})
	}).Children(), "comp-param-defa-value")

	if compParamDefaValue == nil {
		return ""
	}

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
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}
	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(node), "local-comp-def")
	if localCompDef != nil {
		return false
	}
	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(node), "global-comp-def")
	if globalCompDef == nil {
		return false
	}
	return true
}

func (p *paramRefInGlobalCompDef) Render() string {

	paramRefName := p.paramRefName()

	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "global-comp-def")
	globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")

	compParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")

	compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		if strings.TrimSpace(string(compParamName.Raw())) == paramRefName {
			return true
		}
		return false
	})

	compCall := ast.FindNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgs != nil {
		compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
			argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
			if strings.TrimSpace(string(argName.Raw())) == paramRefName {
				return true
			}
			return false
		})
		if compCallArg != nil {
			argValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type").Children(), func(node ast.Node) bool {
				return ast.IsRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg"})
			}).Children(), "comp-call-arg-value")
			return strings.TrimSpace(string(argValue.Raw()))
		}
	}

	compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(compParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-call-param", "comp-bool-param"})
	}).Children(), "comp-param-defa-value")

	return strings.TrimSpace(string(compParamDefaValue.Raw()))
}
