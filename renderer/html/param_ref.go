package html

import (
	"html"
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

func renderCompParamCall(r *renderer, rn renderableNode, paramRefName string) string {
	target := resolveParamFromAncestorsTarget(paramRefName, getAncestorsByInvoker(rn), r)
	if target.name == "" {
		return ""
	}
	if isCompTargetInInvokerChain(r, rn, target.name) {
		return ""
	}

	inlineCall := isInlineCompParamRef(rn.Node())

	localCompDefSrc := target.scope
	if localCompDefSrc == nil {
		localCompDefSrc = localCompSourceFromNode(rn.Node(), r.root)
	}

	localCompDef := r.findLocalCompDef(localCompDefSrc, target.name)
	if localCompDef == nil {
		currentGlobalCompDef := ast.FindNode(ast.GetAncestors(rn.Node()), func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "global-comp-def")
		})
		if currentGlobalCompDef != nil && currentGlobalCompDef != localCompDefSrc {
			localCompDef = r.findLocalCompDef(currentGlobalCompDef, target.name)
		}
	}
	if localCompDef != nil {
		localCompDefContent := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-content")
		if localCompDefContent == nil {
			return ""
		}
		if inlineCall {
			return renderInlineCompDefContent(r, rn, localCompDefContent)
		}
		rendered := r.renderChildren(rn, localCompDefContent.Children())
		if strings.Contains(rendered, "<compono-error-block>") {
			rendered = strings.ReplaceAll(rendered, "<br>", "</p><p>")
		}
		return rendered
	}

	globalCompDef := r.findGlobalCompDef(target.name)
	if globalCompDef != nil {
		globalCompDefContent := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		if inlineCall {
			return renderInlineCompDefContent(r, rn, globalCompDefContent)
		}
		rendered := r.renderChildren(rn, globalCompDefContent.Children())
		if strings.Contains(rendered, "<compono-error-block>") {
			rendered = strings.ReplaceAll(rendered, "<br>", "</p><p>")
		}
		return rendered
	}

	builtinComp := r.findBuiltinComp(target.name)
	if builtinComp != nil {
		return builtinComp.Render(rn.Invoker(), rn.Node())
	}

	return ""
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
	return localCompDef != nil
}

func (p *paramRefInLocalCompDef) Render() string {
	paramRefName := p.paramRefName()

	localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "local-comp-def")
	localCompDefHead := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
	compParams := ast.FindNodeByRuleName(localCompDefHead.Children(), "comp-params")

	var compParam ast.Node
	if compParams != nil {
		compParam = ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
			compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
			return strings.TrimSpace(string(compParamName.Raw())) == paramRefName
		})
	}

	if compParam != nil {
		if shouldTreatParamRefAsCompCall(compParam, p, p.renderer, paramRefName) {
			return renderCompParamCall(p.renderer, p, paramRefName)
		}

		compCall := ast.FindNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
			return isCompCallLikeNode(node)
		})

		if compCall != nil {
			compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
			if compCallArgs != nil {
				compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
					argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
					return strings.TrimSpace(string(argName.Raw())) == paramRefName
				})
				if compCallArg != nil {
					return resolveCompCallArgValue(compCallArg, getAncestorsByInvoker(p), compCall, p.renderer)
				}
			}
		}

		compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
		if compParamType == nil {
			return ""
		}
		compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(compParamType.Children(), func(node ast.Node) bool {
			return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
		}).Children(), "comp-param-defa-value")

		if compParamDefaValue == nil {
			return ""
		}

		return html.EscapeString(strings.TrimSpace(string(compParamDefaValue.Raw())))
	}

	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "global-comp-def")
	if globalCompDef == nil {
		return ""
	}

	globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")
	if globalCompDefHead == nil {
		return ""
	}

	globalCompParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")
	if globalCompParams == nil {
		return ""
	}

	globalCompParam := ast.FindNode(globalCompParams.Children(), func(cp ast.Node) bool {
		compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		return strings.TrimSpace(string(compParamName.Raw())) == paramRefName
	})

	if globalCompParam == nil {
		return ""
	}

	for _, anc := range getAncestorsByInvoker(p) {
		if !isCompCallLikeNode(anc) {
			continue
		}
		compCallArgs := ast.FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs == nil {
			continue
		}
		compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
			argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
			return strings.TrimSpace(string(argName.Raw())) == paramRefName
		})
		if compCallArg != nil {
			return resolveCompCallArgValue(compCallArg, getAncestorsByInvoker(p), anc, p.renderer)
		}
	}

	compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(globalCompParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
	}).Children(), "comp-param-defa-value")

	if compParamDefaValue == nil {
		return ""
	}

	return html.EscapeString(strings.TrimSpace(string(compParamDefaValue.Raw())))
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
	return globalCompDef != nil
}

func (p *paramRefInGlobalCompDef) Render() string {
	paramRefName := p.paramRefName()

	globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(p.Node()), "global-comp-def")
	globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")

	compParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")

	compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
		compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		return strings.TrimSpace(string(compParamName.Raw())) == paramRefName
	})
	if compParam == nil {
		return ""
	}
	if shouldTreatParamRefAsCompCall(compParam, p, p.renderer, paramRefName) {
		return renderCompParamCall(p.renderer, p, paramRefName)
	}

	compCall := ast.FindNode(getAncestorsByInvoker(p), func(node ast.Node) bool {
		return isCompCallLikeNode(node)
	})

	if compCall != nil {
		compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
				argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
				return strings.TrimSpace(string(argName.Raw())) == paramRefName
			})
			if compCallArg != nil {
				return resolveCompCallArgValue(compCallArg, getAncestorsByInvoker(p), compCall, p.renderer)
			}
		}
	}

	compParamDefaValue := ast.FindNodeByRuleName(ast.FindNode(ast.FindNodeByRuleName(compParam.Children(), "comp-param-type").Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
	}).Children(), "comp-param-defa-value")

	if compParamDefaValue == nil {
		return ""
	}

	return html.EscapeString(strings.TrimSpace(string(compParamDefaValue.Raw())))
}

func resolveCompCallArgValue(compCallArg ast.Node, invokerAncestors []ast.Node, currentCompCall ast.Node, r ...*renderer) string {
	compCallArgType := ast.FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	if compCallArgType == nil {
		return ""
	}
	argTypeNode := ast.FindNode(compCallArgType.Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg", "comp-call-param-arg", "comp-call-comp-arg"})
	})
	if argTypeNode == nil {
		return ""
	}

	argValue := ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-arg-value")
	if argValue == nil {
		return ""
	}

	if ast.IsRuleName(argTypeNode, "comp-call-param-arg") {
		referencedParamName := strings.TrimSpace(string(argValue.Raw()))
		var remainingAncestors []ast.Node
		for i, anc := range invokerAncestors {
			if anc == currentCompCall {
				remainingAncestors = invokerAncestors[i+1:]
				break
			}
		}
		var rend *renderer
		if len(r) > 0 {
			rend = r[0]
		}
		return resolveParamFromAncestors(referencedParamName, remainingAncestors, rend)
	}

	return html.EscapeString(strings.TrimSpace(string(argValue.Raw())))
}

func resolveParamFromAncestors(paramName string, invokerAncestors []ast.Node, r *renderer) string {
	for i, anc := range invokerAncestors {
		if !isCompCallLikeNode(anc) {
			continue
		}

		compCallArgs := ast.FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := ast.FindNode(compCallArgs.Children(), func(cca ast.Node) bool {
				argName := ast.FindNodeByRuleName(cca.Children(), "comp-call-arg-name")
				return strings.TrimSpace(string(argName.Raw())) == paramName
			})

			if compCallArg != nil {
				return resolveCompCallArgValue(compCallArg, invokerAncestors, anc, r)
			}
		}

		if r != nil {
			if ast.IsRuleName(anc, "param-ref") {
				paramRefName := getParamRefNameStr(anc)
				if paramRefName != "" {
					target := resolveParamFromAncestorsTarget(paramRefName, invokerAncestors[i+1:], r)
					if target.name != "" {
						compDef := r.findLocalCompDef(target.scope, target.name)
						if compDef == nil {
							compDef = r.findGlobalCompDef(target.name)
						}
						if compDef != nil {
							if defaValue := getCompParamDefault(compDef, paramName); defaValue != "" {
								return html.EscapeString(defaValue)
							}
						}
					}
				}
				continue
			}

			val := resolveParamDefaultFromCompCall(anc, paramName, r)
			if val != "" {
				return val
			}
		}
	}
	return ""
}

func resolveParamDefaultFromCompCall(compCallNode ast.Node, paramName string, r *renderer) string {
	compDef := findCompDefFromCompCall(compCallNode, r)
	if compDef == nil {
		return ""
	}

	var compDefHead ast.Node
	compDefHead = ast.FindNodeByRuleName(compDef.Children(), "local-comp-def-head")
	if compDefHead == nil {
		compDefHead = ast.FindNodeByRuleName(compDef.Children(), "global-comp-def-head")
	}
	if compDefHead == nil {
		return ""
	}

	compParams := ast.FindNodeByRuleName(compDefHead.Children(), "comp-params")
	if compParams == nil {
		return ""
	}

	compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
		cpName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
		return cpName != nil && strings.TrimSpace(string(cpName.Raw())) == paramName
	})
	if compParam == nil {
		return ""
	}

	compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
	if compParamType == nil {
		return ""
	}

	typeNode := ast.FindNode(compParamType.Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param"})
	})
	if typeNode == nil {
		return ""
	}

	defaValue := ast.FindNodeByRuleName(typeNode.Children(), "comp-param-defa-value")
	if defaValue == nil {
		return ""
	}

	return html.EscapeString(strings.TrimSpace(string(defaValue.Raw())))
}

func findCompDefFromCompCall(compCallNode ast.Node, r *renderer) ast.Node {
	compCallNameNode := ast.FindNodeByRuleName(compCallNode.Children(), "comp-call-name")
	if compCallNameNode == nil {
		return nil
	}
	compName := strings.TrimSpace(string(compCallNameNode.Raw()))

	globalCompDefAnc := ast.FindNode(ast.GetAncestors(compCallNode), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := r.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := r.findLocalCompDef(localCompDefSrc, compName)
	if localCompDef != nil {
		return localCompDef
	}

	return r.findGlobalCompDef(compName)
}

func getParamRefNameStr(node ast.Node) string {
	refNameNode := ast.FindNodeByRuleName(node.Children(), "param-ref-name")
	if refNameNode != nil {
		return strings.TrimSpace(string(refNameNode.Raw()))
	}
	return ""
}

func renderInlineCompDefContent(r *renderer, invoker renderableNode, compDefContent ast.Node) string {
	childCount := len(compDefContent.Children())
	if childCount == 0 {
		return ""
	}
	p := ast.FindNodeByRuleName(compDefContent.Children(), "p")
	if p == nil {
		return ""
	}
	pContent := ast.FindNodeByRuleName(p.Children(), "p-content")
	if pContent == nil {
		return ""
	}

	return r.renderChildren(invoker, pContent.Children())
}

func isInlineCompParamRef(node ast.Node) bool {
	pContent := ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "p-content")
	})
	if pContent != nil {
		return !isStandaloneParamRefInParagraph(node, pContent)
	}

	return ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{
			"h1-content",
			"h2-content",
			"h3-content",
			"h4-content",
			"h5-content",
			"h6-content",
			"em-content",
			"strong-content",
			"link-text",
		})
	}) != nil
}

func isStandaloneParamRefInParagraph(paramRef ast.Node, pContent ast.Node) bool {
	for _, child := range pContent.Children() {
		if child == paramRef {
			continue
		}

		if ast.IsRuleName(child, "plain") {
			if strings.TrimSpace(string(child.Raw())) == "" {
				continue
			}
		}

		return false
	}

	return true
}

func standaloneCompParamRefInParagraph(pContent ast.Node) ast.Node {
	paramRef := ast.FindNode(pContent.Children(), func(node ast.Node) bool {
		return ast.IsRuleName(node, "param-ref")
	})
	if paramRef == nil {
		return nil
	}
	if !isStandaloneParamRefInParagraph(paramRef, pContent) {
		return nil
	}
	compParam := findParamDefByRef(paramRef)
	if compParam == nil {
		return nil
	}
	paramType := ast.GetTypeFromCompParam(compParam)
	if paramType != "comp" && paramType != "" {
		return nil
	}
	return paramRef
}

func findParamDefByRef(paramRef ast.Node) ast.Node {
	paramRefName := getParamRefNameStr(paramRef)
	if paramRefName == "" {
		return nil
	}

	localCompDef := ast.FindNode(ast.GetAncestors(paramRef), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "local-comp-def")
	})
	if localCompDef != nil {
		if compParam := findCompParamByName(localCompDef, paramRefName); compParam != nil {
			return compParam
		}
	}

	globalCompDef := ast.FindNode(ast.GetAncestors(paramRef), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})
	if globalCompDef != nil {
		return findCompParamByName(globalCompDef, paramRefName)
	}

	return nil
}

func findCompParamByName(compDef ast.Node, name string) ast.Node {
	compParams := ast.GetCompParamsFromCompDef(compDef)
	if len(compParams) == 0 {
		return nil
	}

	return ast.FindNode(compParams, func(cp ast.Node) bool {
		return ast.GetParamNameFromCompParam(cp) == name
	})
}

func isCompTargetInInvokerChain(r *renderer, rn renderableNode, targetName string) bool {
	ancestors := getAncestorsByInvoker(rn)
	for i, anc := range ancestors {
		if ast.IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call"}) {
			compCallName := ast.FindNodeByRuleName(anc.Children(), "comp-call-name")
			if compCallName != nil && strings.TrimSpace(string(compCallName.Raw())) == targetName {
				return true
			}
			continue
		}

		if !ast.IsRuleName(anc, "param-ref") {
			continue
		}

		paramRefName := getParamRefNameStr(anc)
		if paramRefName == "" {
			continue
		}
		resolved := resolveParamFromAncestorsTarget(paramRefName, ancestors[i+1:], r)
		if resolved.name == targetName {
			return true
		}
	}

	return false
}

func shouldTreatParamRefAsCompCall(compParam ast.Node, rn renderableNode, r *renderer, paramRefName string) bool {
	paramType := ast.GetTypeFromCompParam(compParam)
	if paramType == "comp" {
		return true
	}
	if paramType != "" {
		return false
	}

	if rn != nil {
		if ast.FindNodeByRuleName(rn.Node().Children(), "comp-call-args") != nil {
			return true
		}
	}

	if rn == nil || r == nil {
		return false
	}

	target := resolveParamFromAncestorsTarget(paramRefName, getAncestorsByInvoker(rn), r)
	if target.name == "" {
		return false
	}
	if target.name == strings.ToUpper(target.name) {
		return true
	}

	if r.findBuiltinCompDef(target.name) != nil {
		return true
	}
	if r.findLocalCompDef(target.scope, target.name) != nil {
		return true
	}
	return r.findGlobalCompDef(target.name) != nil
}
