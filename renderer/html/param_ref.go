package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/errwrap"
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
	resolved := resolveParamRefValue(rn, r, paramRefName)
	if resolved.Type != "comp" || resolved.Raw == "" {
		return ""
	}
	target := resolvedCompTarget{name: resolved.Raw, scope: resolved.Scope}
	if isCompTargetInInvokerChain(r, rn, target.name) {
		return ""
	}

	inlineCall := isInlineCompParamRef(rn.Node())

	localCompDefSrc := target.scope
	if localCompDefSrc == nil {
		localCompDefSrc = ast.GetLocalCompSourceFromNode(rn.Node(), r.root)
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
			if title, message, ok := errwrap.InlineCompValueError(target.name, localCompDef); ok {
				return r.recordDiagnostic(rn.Node(), title, message, false)
			}
		}
		r.pushLocalFrame(target.name, rn.Node())
		defer r.popFrame()
		if inlineCall {
			return renderInlineCompDefContent(r, rn, localCompDefContent)
		}
		return r.renderChildren(rn, localCompDefContent.Children())
	}

	globalCompDef := r.findGlobalCompDef(target.scope, target.name)
	if globalCompDef != nil {
		globalCompDefContent := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-content")
		if globalCompDefContent == nil {
			return ""
		}
		if inlineCall {
			if title, message, ok := errwrap.InlineCompValueError(target.name, globalCompDef); ok {
				return r.recordDiagnostic(rn.Node(), title, message, false)
			}
		}
		r.pushGlobalFrame(target.name, globalCompDef, rn.Node())
		defer r.popFrame()
		if inlineCall {
			return renderInlineCompDefContent(r, rn, globalCompDefContent)
		}
		return r.renderChildren(rn, globalCompDefContent.Children())
	}

	builtinComp := r.findBuiltinComp(target.name)
	if builtinComp != nil {
		return builtinComp.Render(rn.Invoker(), rn.Node())
	}

	return ""
}

type resolvedCompTarget struct {
	name  string
	scope ast.Node
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

		return renderParamRefUnit(paramRefName, p, p.renderer)
	}

	return ""
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

	return renderParamRefUnit(paramRefName, p, p.renderer)
}

func renderResolvedValue(value ast.ResolvedValue) string {
	if value.IsZero() || value.Type == "array" || value.Type == "record" {
		return ""
	}

	return html.EscapeString(strings.TrimSpace(value.Raw))
}

// renderParamRefUnit renders a parameter reference unit with a value. An
// error of the resolved value drops the unit.
func renderParamRefUnit(paramName string, rn renderableNode, r *renderer) string {
	accessors := ast.GetParamRefAccessors(rn.Node())
	value := ast.ResolveParamFromAncestors(r.root, paramName, nil, getAncestorsByInvoker(rn))
	if title, message, ok := errwrap.ParamRefValueError(paramName, value, accessors); ok {
		return r.recordDiagnostic(rn.Node(), title, message, false)
	}
	return renderResolvedValue(ast.ApplyAccessors(value, accessors))
}

func resolveParamRefValue(rn renderableNode, r *renderer, paramName string) ast.ResolvedValue {
	return ast.ResolveParamFromAncestors(r.root, paramName, ast.GetParamRefAccessors(rn.Node()), getAncestorsByInvoker(rn))
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
		if ast.FindNodeByRuleName(pContent.Children(), "soft-break") != nil {
			return !isStandaloneParamRefOnLine(node, pContent)
		}
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

		if ast.IsRuleName(child, "soft-break") {
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

func isStandaloneParamRefOnLine(paramRef ast.Node, pContent ast.Node) bool {
	line := []ast.Node{}
	for _, child := range pContent.Children() {
		if ast.IsRuleName(child, "soft-break") {
			if containsNode(line, paramRef) {
				return isStandaloneWithinNodes(paramRef, line)
			}
			line = []ast.Node{}
			continue
		}
		line = append(line, child)
	}

	if containsNode(line, paramRef) {
		return isStandaloneWithinNodes(paramRef, line)
	}

	return false
}

func containsNode(nodes []ast.Node, target ast.Node) bool {
	for _, node := range nodes {
		if node == target {
			return true
		}
	}
	return false
}

func isStandaloneWithinNodes(paramRef ast.Node, nodes []ast.Node) bool {
	for _, child := range nodes {
		if child == paramRef {
			continue
		}
		if ast.IsRuleName(child, "plain") && strings.TrimSpace(string(child.Raw())) == "" {
			continue
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
	return paramRef
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

		resolved := ast.ResolveFrameCompValue(r.root, anc, ancestors[i+1:])
		if resolved.Type == "comp" && resolved.Raw == targetName {
			return true
		}
	}

	return false
}

func shouldTreatParamRefAsCompCall(compParam ast.Node, rn renderableNode, r *renderer, paramRefName string) bool {
	if rn != nil && r != nil && len(ast.GetParamRefAccessors(rn.Node())) > 0 {
		return resolveParamRefValue(rn, r, paramRefName).Type == "comp"
	}

	paramType := ast.GetTypeFromCompParam(compParam)
	if paramType == "context" {
		paramType = resolveParamRefValue(rn, r, paramRefName).Type
	}
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

	return resolveParamRefValue(rn, r, paramRefName).Type == "comp"
}
