package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
)

type ErrorWrapper interface {
	Wrap(ast.Node)
}

func DefaultErrorWrapper() ErrorWrapper {
	return &errorWrapper{
		wrapRules: wrapRules(),
	}
}

type errorWrapper struct {
	wrapRules []wrapRule
}

func (ew *errorWrapper) Wrap(root ast.Node) {
	ctx := &wrapContext{
		root:             root,
		compCallChains:   ew.getCompCallChains(root),
		callReplacements: ew.getCallReplacements(root),
	}

	ew.scanAndWrap(ctx, root)
}

func (ew *errorWrapper) scanAndWrap(ctx *wrapContext, node ast.Node) {
	if replacement, ok := ctx.callReplacements[node]; ok {
		ew.replaceNode(node, replacement)
		return
	}

	if ew.wrap(ctx, node) {
		return
	}

	for _, child := range node.Children() {
		ew.scanAndWrap(ctx, child)
	}
}

func (ew *errorWrapper) wrap(ctx *wrapContext, node ast.Node) (wrapped bool) {
Outer:
	for _, wr := range ew.wrapRules {
		for _, cond := range wr.conditions {
			if !cond(ctx, node) {
				continue Outer
			}
		}
		ew.wrapWithErr(node, wr.title(ctx, node), wr.message(ctx, node), wr.block(ctx, node))
		return true
	}

	return false
}

func (ew *errorWrapper) getCompCallChains(root ast.Node) [][]ast.Node {
	rootContent := ast.FindNodeByRuleName(root.Children(), "root-content")
	compCalls := ast.FilterNodesInTree(rootContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	chains := [][]ast.Node{}

	for _, compCall := range compCalls {
		chain := []ast.Node{}
		addLinkToChain(root, &chain, compCall)
		chains = append(chains, chain)
	}

	return chains
}

func addLinkToChain(root ast.Node, chain *[]ast.Node, compCall ast.Node) {
	stop := false
	for _, existing := range *chain {
		if existing == compCall {
			stop = true
		}
	}

	*chain = append(*chain, compCall)
	if stop {
		return
	}

	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return
	}

	compDef := findCompDef(root, compCall, compCallName)
	if compDef == nil {
		return
	}

	compCalls := ast.FilterNodesInTree(compDef, func(child ast.Node) bool {
		return ast.IsRuleNameOneOf(child, []string{"block-comp-call", "inline-comp-call"})
	})

	for _, cc := range compCalls {
		addLinkToChain(root, chain, cc)
	}
}

func (ew *errorWrapper) wrapWithErr(self ast.Node, title, msg string, block bool) {
	var errNode ast.Node
	if block {
		errNode = ew.createBlockError(self, title, msg)
	} else {
		errNode = ew.createInlineError(self, title, msg)
	}

	self.SetRule(errNode.Rule())
	self.SetChildren(errNode.Children())
	self.SetRaw(errNode.Raw())
}

func (ew *errorWrapper) replaceNode(self ast.Node, replacement ast.Node) {
	self.SetRule(replacement.Rule())
	self.SetChildren(replacement.Children())
	self.SetRaw(replacement.Raw())
	for _, child := range self.Children() {
		child.SetParent(self)
	}
}

func (ew *errorWrapper) createBlockError(node ast.Node, title, msg string) ast.Node {
	return ew.createError("block-error", node, title, msg)
}

func (ew *errorWrapper) createInlineError(node ast.Node, title, msg string) ast.Node {
	return ew.createError("inline-error", node, title, msg)
}

func (ew *errorWrapper) createError(errRuleName string, node ast.Node, title, msg string) ast.Node {
	err := rule.NewDynamic(errRuleName)
	errTitle := rule.NewDynamic("error-title")
	errMsg := rule.NewDynamic("error-message")
	self := rule.NewDynamic("self")

	errNode := ast.DefaultEmptyNode()
	errNode.SetRule(err)

	errTitleNode := ast.DefaultEmptyNode()
	errTitleNode.SetRule(errTitle)
	errTitleNode.SetParent(errNode)
	errTitleNode.SetRaw([]byte(title))

	errMsgNode := ast.DefaultEmptyNode()
	errMsgNode.SetRule(errMsg)
	errMsgNode.SetParent(errNode)
	errMsgNode.SetRaw([]byte(msg))

	selfNode := ast.DefaultEmptyNode()
	selfNode.SetRule(self)
	selfNode.SetParent(errNode)
	selfNode.SetChildren(node.Children())

	errNode.SetChildren([]ast.Node{
		errTitleNode,
		errMsgNode,
		selfNode,
	})

	return errNode
}

func (ew *errorWrapper) getCallReplacements(root ast.Node) map[ast.Node]ast.Node {
	result := map[ast.Node]ast.Node{}

	rootContent := ast.FindNodeByRuleName(root.Children(), "root-content")
	if rootContent == nil {
		return result
	}

	rootCompCalls := ast.FilterNodesInTree(rootContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	for _, compCall := range rootCompCalls {
		replacement := ew.getReplacementForCompCall(root, compCall)
		if replacement != nil {
			result[compCall] = replacement
		}

		for node, replacement := range ew.getInlineParamRefReplacements(root, compCall) {
			result[node] = replacement
		}
	}

	return result
}

func (ew *errorWrapper) getReplacementForCompCall(root ast.Node, compCall ast.Node) ast.Node {
	compDef := findCompDef(root, compCall, getCompCallNameStr(compCall))
	if compDef == nil {
		return nil
	}

	content := getCompDefContent(compDef)
	if content == nil {
		return nil
	}

	paramValues := resolveCompCallParamValues(root, compDef, compCall)
	if len(paramValues) == 0 {
		return nil
	}

	for _, child := range content.Children() {
		if !ast.IsRuleName(child, "p") {
			continue
		}

		replacement := ew.getParagraphReplacement(compDef, child, paramValues)
		if replacement != nil {
			return replacement
		}
	}

	return nil
}

func (ew *errorWrapper) getInlineParamRefReplacements(root ast.Node, compCall ast.Node) map[ast.Node]ast.Node {
	result := map[ast.Node]ast.Node{}

	compDef := findCompDef(root, compCall, getCompCallNameStr(compCall))
	if compDef == nil {
		return result
	}

	content := getCompDefContent(compDef)
	if content == nil {
		return result
	}

	paramValues := resolveCompCallParamValues(root, compDef, compCall)
	if len(paramValues) == 0 {
		return result
	}

	inlineParamRefs := ast.FilterNodesInTree(content, func(node ast.Node) bool {
		if !ast.IsRuleName(node, "param-ref") || hasCompCallArgsNode(node) || !isInlineParamRefNode(node) || !canRenderParamRefAsValue(compDef, node) {
			return false
		}

		for _, accessor := range ast.GetParamRefAccessors(node) {
			if accessor.Kind == "key" {
				return true
			}
		}

		return false
	})

	for _, paramRef := range inlineParamRefs {
		paramName := getParamRefNameStr(paramRef)
		resolved, ok := paramValues[paramName]
		if !ok {
			continue
		}

		indexed, _ := ast.ApplyAccessorsDetailed(resolved, ast.GetParamRefAccessors(paramRef))
		if indexed.Type != "comp" || indexed.Raw == "" {
			continue
		}

		targetCompDef := findCompDef(root, compCall, indexed.Raw)
		if targetCompDef == nil || !isBlockComponent(targetCompDef) {
			continue
		}

		result[paramRef] = ew.createInlineError(paramRef, "Invalid component usage", "The component **"+indexed.Raw+"** is a block component and cannot be used inline.")
	}

	return result
}

func (ew *errorWrapper) getParagraphReplacement(compDef ast.Node, p ast.Node, values map[string]ast.ResolvedValue) ast.Node {
	pContent := ast.FindNodeByRuleName(p.Children(), "p-content")
	if pContent == nil {
		return nil
	}

	hasSoftBreak := ast.FindNodeByRuleName(pContent.Children(), "soft-break") != nil

	for _, child := range pContent.Children() {
		if !ast.IsRuleName(child, "param-ref") || hasCompCallArgsNode(child) || !canRenderParamRefAsValue(compDef, child) {
			continue
		}

		paramName := getParamRefNameStr(child)
		resolved, ok := values[paramName]
		if !ok {
			continue
		}

		accessors := ast.GetParamRefAccessors(child)
		indexed, accessErr := ast.ApplyAccessorsDetailed(resolved, accessors)
		if len(accessors) > 0 && indexed.IsZero() && accessErr.Kind == "array_index_out_of_range" {
			if hasSoftBreak && isStandaloneParamRefOnLineInParagraph(child, pContent) {
				return buildParagraphErrorNode(ew.createInlineError(child, "Array index out of range", "The index used for parameter **"+paramName+"** is out of range."))
			}
			return cloneParagraphWithReplacement(p, child, ew.createInlineError(child, "Array index out of range", "The index used for parameter **"+paramName+"** is out of range."))
		}

		if len(accessors) == 0 && indexed.Type == "array" {
			return cloneParagraphWithReplacement(p, child, ew.createInlineError(child, "Invalid parameter usage", "The parameter **"+paramName+"** is an array and cannot be rendered directly."))
		}

		if len(accessors) == 0 && indexed.Type == "record" {
			return cloneParagraphWithReplacement(p, child, ew.createInlineError(child, "Invalid parameter usage", "The parameter **"+paramName+"** is a record and cannot be rendered directly."))
		}

		if accessErr.Kind == "unknown_record_key" {
			if hasSoftBreak && isStandaloneParamRefOnLineInParagraph(child, pContent) {
				return buildParagraphErrorNode(ew.createInlineError(child, "Unknown record key", "The key **"+accessErr.Key+"** is not defined in this record."))
			}
			return cloneParagraphWithReplacement(p, child, ew.createInlineError(child, "Unknown record key", "The key **"+accessErr.Key+"** is not defined in this record."))
		}
	}

	return nil
}

func canRenderParamRefAsValue(compDef ast.Node, paramRef ast.Node) bool {
	paramName := getParamRefNameStr(paramRef)
	if paramName == "" {
		return false
	}

	for _, info := range getCompDefParamInfos(compDef) {
		if info.name != paramName {
			continue
		}
		return info.typ != "comp"
	}

	return true
}

func resolveCompCallParamValues(root ast.Node, compDef ast.Node, compCall ast.Node) map[string]ast.ResolvedValue {
	values := map[string]ast.ResolvedValue{}

	for _, compParam := range ast.GetCompParamsFromCompDef(compDef) {
		name := ast.GetParamNameFromCompParam(compParam)
		if name == "" {
			continue
		}
		values[name] = ast.ResolveCompParamDefaultFromCompDef(root, compDef, name)
	}

	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		name := ast.GetArgNameFromCompCallArg(arg)
		if name == "" {
			continue
		}
		values[name] = ast.ResolveCompCallArgValue(root, arg, ast.GetAncestors(compCall), compCall)
	}

	return values
}

func buildParagraphErrorNode(inlineError ast.Node) ast.Node {
	p := ast.DefaultEmptyNode()
	p.SetRule(rule.NewDynamic("p"))

	pContent := ast.DefaultEmptyNode()
	pContent.SetRule(rule.NewDynamic("p-content"))
	pContent.SetParent(p)

	inlineError.SetParent(pContent)
	pContent.SetChildren([]ast.Node{inlineError})
	p.SetChildren([]ast.Node{pContent})

	return p
}

func cloneParagraphWithReplacement(srcP ast.Node, target ast.Node, replacement ast.Node) ast.Node {
	cloned := cloneNode(srcP, map[ast.Node]ast.Node{
		target: replacement,
	})
	return cloned
}

func cloneNode(src ast.Node, replacements map[ast.Node]ast.Node) ast.Node {
	if replacement, ok := replacements[src]; ok {
		return cloneNode(replacement, nil)
	}

	node := ast.DefaultEmptyNode()
	node.SetRule(rule.NewDynamic(src.Rule().Name()))
	node.SetRaw(src.Raw())

	children := make([]ast.Node, 0, len(src.Children()))
	for _, child := range src.Children() {
		clonedChild := cloneNode(child, replacements)
		clonedChild.SetParent(node)
		children = append(children, clonedChild)
	}
	node.SetChildren(children)

	return node
}

func isStandaloneParamRefOnLineInParagraph(paramRef ast.Node, pContent ast.Node) bool {
	line := []ast.Node{}
	for _, child := range pContent.Children() {
		if ast.IsRuleName(child, "soft-break") {
			if lineContainsNode(line, paramRef) {
				return isStandaloneLine(paramRef, line)
			}
			line = []ast.Node{}
			continue
		}
		line = append(line, child)
	}

	if lineContainsNode(line, paramRef) {
		return isStandaloneLine(paramRef, line)
	}

	return false
}

func lineContainsNode(nodes []ast.Node, target ast.Node) bool {
	for _, node := range nodes {
		if node == target {
			return true
		}
	}
	return false
}

func isStandaloneLine(paramRef ast.Node, line []ast.Node) bool {
	for _, child := range line {
		if child == paramRef {
			continue
		}
		if ast.IsRuleName(child, "plain") && string(child.Raw()) == "" {
			continue
		}
		if ast.IsRuleName(child, "plain") && len(string(child.Raw())) > 0 {
			if len([]rune(string(child.Raw()))) == len([]rune(strings.TrimSpace(string(child.Raw())))) {
				return false
			}
			continue
		}
		return false
	}
	return true
}
