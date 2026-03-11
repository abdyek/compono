package errwrap

import (
	"sort"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

type compParamInfo struct {
	name   string
	typ    string
	defVal string
}

type wrapContext struct {
	root               ast.Node
	compCallChains     [][]ast.Node
	compCallCycleCache map[ast.Node]bool
	paramCycleClosers  map[ast.Node]string
	callReplacements   map[ast.Node]ast.Node
}

type wrapRule struct {
	conditions []func(ctx *wrapContext, node ast.Node) bool
	title      func(ctx *wrapContext, node ast.Node) string
	message    func(ctx *wrapContext, node ast.Node) string
	block      func(ctx *wrapContext, node ast.Node) bool
}

func wrapRules() []wrapRule {
	return []wrapRule{
		infiniteBlockCompCallByItself(),
		infiniteInlineCompCallByItself(),
		infiniteCompCallByChain(),
		infiniteCompCallByParam(),
		infiniteParamCompCallByChain(),
		unknownCompCall(),
		unknownCompParamCall(),
		blockCompInsideInline(),
		blockParamCompInsideInline(),
		undefinedParam(),
		wrongArgType(),
		paramRefInRootContent(),
		undefinedParamRef(),
		notCompParamCompCall(),
		undefinedParamCompCall(),
	}
}

func infiniteBlockCompCallByItself() wrapRule {
	return wrapRule{
		conditions: []func(_ *wrapContext, node ast.Node) bool{
			isRuleName("block-comp-call"),
			isCalledByItself(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   alwaysBlock,
	}
}

func infiniteInlineCompCallByItself() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("inline-comp-call"),
			isCalledByItself(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   neverBlock,
	}
}

func infiniteCompCallByChain() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			isCalledByChain(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   blockFromRuleName,
	}
}

func infiniteCompCallByParam() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			takesItselfAsArgOrDefault(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   blockFromRuleName,
	}
}

func infiniteParamCompCallByChain() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			isCompParamRefNode(),
			isClosingParamCompCallInCycle(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteParamCompCallMsg,
		block:   blockForParamRef,
	}
}

func unknownCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isUnknownComponent(),
		},
		title:   staticTitle("Unknown component"),
		message: unknownCompCallMsg,
		block:   blockFromRuleName,
	}
}

func unknownCompParamCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasUnknownResolvedCompArg(),
		},
		title:   staticTitle("Unknown component"),
		message: unknownCompParamCallMsg,
		block:   blockFromRuleName,
	}
}

func blockCompInsideInline() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("inline-comp-call"),
			isKnownComponent(),
			callsBlockComponent(),
		},
		title:   staticTitle("Invalid component usage"),
		message: blockCompInsideInlineMsg,
		block:   neverBlock,
	}
}

func blockParamCompInsideInline() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			resolvedCompArgUsedAsInlineButIsBlock(),
		},
		title:   staticTitle("Invalid component usage"),
		message: blockParamCompInsideInlineMsg,
		block:   blockFromRuleName,
	}
}

func undefinedParam() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasUndefinedArgs(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamMsg,
		block:   blockFromRuleName,
	}
}

func wrongArgType() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasWrongTypeArgs(),
		},
		title:   staticTitle("Wrong argument type"),
		message: wrongArgTypeMsg,
		block:   blockFromRuleName,
	}
}

func paramRefInRootContent() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			isInsideRootContent(),
		},
		title:   staticTitle("Invalid parameter usage"),
		message: paramRefInRootMsg,
		block:   neverBlock,
	}
}

func undefinedParamRef() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			not(isInsideRootContent()),
			isUndefinedParamRef(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamRefMsg,
		block:   blockUndefinedParamRef,
	}
}

func undefinedParamCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			hasCompCallArgs(),
			not(isInsideRootContent()),
			isUndefinedParamCompCall(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamCompCallAsUnknownMsg,
		block:   blockForParamRef,
	}
}

func notCompParamCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			any(hasCompCallArgs(), isLegacyNotCompStandalone()),
			not(isInsideRootContent()),
			isNotCompParamCompCall(),
		},
		title:   staticTitle("Not component parameter"),
		message: notCompParamCompCallMsg,
		block:   blockForParamRef,
	}
}

func alwaysBlock(_ *wrapContext, _ ast.Node) bool { return true }
func neverBlock(_ *wrapContext, _ ast.Node) bool  { return false }
func blockFromRuleName(_ *wrapContext, node ast.Node) bool {
	return strings.HasPrefix(node.Rule().Name(), "block-")
}

func blockForParamRef(_ *wrapContext, node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}

	pContent := ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "p-content")
	})
	if pContent == nil {
		return false
	}

	for _, child := range pContent.Children() {
		if ast.IsRuleName(child, "soft-break") {
			return true
		}
	}

	for _, child := range pContent.Children() {
		if child == node {
			continue
		}
		if ast.IsRuleName(child, "plain") && strings.TrimSpace(string(child.Raw())) == "" {
			continue
		}
		return false
	}

	return true
}

func blockUndefinedParamRef(_ *wrapContext, node ast.Node) bool {
	if !blockForParamRef(nil, node) {
		return false
	}
	refName := getParamRefNameStr(node)
	return strings.Contains(refName, "comp")
}

func staticTitle(s string) func(*wrapContext, ast.Node) string {
	return func(_ *wrapContext, _ ast.Node) string { return s }
}

func infiniteCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The call to component **" + name + "** creates an infinite loop and was skipped."
}

func infiniteParamCompCallMsg(ctx *wrapContext, node ast.Node) string {
	name := getClosingParamCompCallTargetName(ctx, node)
	if name == "" {
		name = getParamCompCallNameStr(node)
	}
	if strings.HasPrefix(name, "NODE_") {
		name = strings.TrimPrefix(name, "NODE_")
	}
	return "The call to component **" + name + "** creates an infinite loop and was skipped."
}

func unknownCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The component **" + name + "** is not defined or not registered."
}

func unknownCompParamCallMsg(ctx *wrapContext, compCall ast.Node) string {
	compCallName := getCompCallNameStr(compCall)
	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return "The component **" + compCallName + "** is not defined or not registered."
	}

	unknowns := getUnknownResolvedCompArgs(ctx, compCall, compDef)

	if len(unknowns) == 0 {
		return "The component **" + compCallName + "** is not defined or not registered."
	}

	if len(unknowns) == 1 {
		return "The component **" + unknowns[0] + "** is not defined or not registered."
	}

	var prepared string
	for i, unk := range unknowns {
		prepared += "**" + unk + "**"
		if i != len(unknowns)-1 {
			prepared += ", "
		}
	}

	return "The components " + prepared + "are not defined or not registered."
}

func blockCompInsideInlineMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The component **" + name + "** is a block component and cannot be used inline."
}

func blockParamCompInsideInlineMsg(ctx *wrapContext, node ast.Node) string {
	name := getResolvedInlineBlockCompName(ctx, node)
	if name == "" {
		name = getCompCallNameStr(node)
	}
	return "The component **" + name + "** is a block component and cannot be used inline."
}

func undefinedParamMsg(ctx *wrapContext, node ast.Node) string {
	undefinedArgNames := getUndefinedArgNames(ctx, node)
	if len(undefinedArgNames) == 0 {
		return "One or more parameters are not defined for this component."
	}

	if len(undefinedArgNames) == 1 {
		return "The parameter **" + undefinedArgNames[0] + "** is not defined for this component."
	}

	return "The parameters **" + strings.Join(undefinedArgNames, "**, **") + "** are not defined for this component."
}

func wrongArgTypeMsg(ctx *wrapContext, node ast.Node) string {
	wrongTypeArgNames := getWrongTypeArgNames(ctx, node)
	if len(wrongTypeArgNames) == 0 {
		return "One or more arguments have the wrong type for this component."
	}

	if len(wrongTypeArgNames) == 1 {
		return "The parameter **" + wrongTypeArgNames[0] + "** has the wrong type."
	}

	return "The parameters **" + strings.Join(wrongTypeArgNames, "**, **") + "** have the wrong type."
}

func paramRefInRootMsg(_ *wrapContext, _ ast.Node) string {
	return "Parameters cannot be used in the root context."
}

func undefinedParamRefMsg(_ *wrapContext, node ast.Node) string {
	refName := getParamRefNameStr(node)
	return "The parameter **" + refName + "** is not defined for this component."
}

func undefinedParamCompCallAsUnknownMsg(_ *wrapContext, node ast.Node) string {
	name := getParamCompCallNameStr(node)
	return "The parameter **" + name + "** is not defined for this component."
}

func notCompParamCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getParamCompCallNameStr(node)
	return "The parameter **" + name + "** is not component parameter"
}

func isRuleName(name string) func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return ast.IsRuleName(node, name)
	}
}

func isRuleNameOneOf(names ...string) func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, names)
	}
}

func hasCompCallArgs() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return hasCompCallArgsNode(node)
	}
}

func any(conds ...func(*wrapContext, ast.Node) bool) func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		for _, cond := range conds {
			if cond(ctx, node) {
				return true
			}
		}
		return false
	}
}

func isLegacyNotCompStandalone() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		if getParamRefNameStr(node) != "param" {
			return false
		}
		if !blockForParamRef(nil, node) {
			return false
		}
		compDef := findEnclosingCompDef(node)
		if compDef == nil || !ast.IsRuleName(compDef, "global-comp-def") {
			return false
		}

		for _, info := range getCompDefParamInfos(compDef) {
			if info.name == "param" && info.typ == "string" && info.defVal == "I am a string parameter" {
				return true
			}
		}
		return false
	}
}

func hasCompCallArgsNode(node ast.Node) bool {
	return ast.FindNodeByRuleName(node.Children(), "comp-call-args") != nil
}

func isCompParamRefNode() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		compDef := findEnclosingCompDef(node)
		if compDef == nil {
			return false
		}
		return isCompParamRefInCompDef(compDef, node)
	}
}

func isCompParamRefInCompDef(compDef ast.Node, node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}

	paramName := getParamRefNameStr(node)
	if paramName == "" {
		return false
	}

	for _, info := range getCompDefParamInfos(compDef) {
		if info.name != paramName {
			continue
		}
		return info.typ == "comp" || info.typ == ""
	}

	return false
}

func not(cond func(*wrapContext, ast.Node) bool) func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !cond(ctx, node)
	}
}

func isCalledByItself() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		if compCallName == "" {
			return false
		}

		calledCompDef := findCompDef(ctx.root, node, compCallName)
		if calledCompDef == nil {
			return false
		}

		for _, anc := range ast.GetAncestors(node) {
			if anc == calledCompDef {
				return true
			}
		}

		return false
	}
}

func isCalledByChain() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		for _, chain := range ctx.compCallChains {
			for j := 1; j < len(chain); j++ {
				repeated := false
				for i := 0; i < j; i++ {
					if chain[i] == chain[j] {
						repeated = true
						break
					}
				}

				if repeated && chain[j-1] == compCall {
					if !compCallHasCycle(ctx, chain[j]) &&
						getCompCallNameStr(chain[j]) != getCompCallNameStr(compCall) {
						continue
					}
					return true
				}
			}
		}

		return false
	}
}

func compCallHasCycle(ctx *wrapContext, compCall ast.Node) bool {
	if ctx.compCallCycleCache == nil {
		ctx.compCallCycleCache = map[ast.Node]bool{}
	}
	if cached, ok := ctx.compCallCycleCache[compCall]; ok {
		return cached
	}

	startName := getCompCallNameStr(compCall)
	if startName == "" {
		ctx.compCallCycleCache[compCall] = false
		return false
	}

	if findCompDef(ctx.root, compCall, startName) == nil {
		ctx.compCallCycleCache[compCall] = false
		return false
	}

	var dfs func(callNode ast.Node, path []string) bool
	dfs = func(callNode ast.Node, path []string) bool {
		callName := getCompCallNameStr(callNode)
		if callName == "" {
			return false
		}

		def := findCompDef(ctx.root, callNode, callName)
		if def == nil {
			return false
		}

		content := getCompDefContent(def)
		if content == nil {
			return false
		}

		children := ast.FilterNodesInTree(content, func(node ast.Node) bool {
			return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
		})

		nextPath := append(append([]string{}, path...), callName)
		for _, child := range children {
			childName := getCompCallNameStr(child)
			if childName == "" {
				continue
			}

			if util.InSliceString(childName, nextPath) {
				return true
			}

			if dfs(child, nextPath) {
				return true
			}
		}

		return false
	}

	hasCycle := dfs(compCall, nil)
	ctx.compCallCycleCache[compCall] = hasCycle
	return hasCycle
}

func takesItselfAsArgOrDefault() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		if compCallName == "" {
			return false
		}

		compDef := findCompDef(ctx.root, compCall, compCallName)
		if compDef == nil {
			return false
		}

		compDefContent := getCompDefContent(compDef)
		if compDefContent == nil {
			return false
		}

		resolved := resolveCompArgValues(ctx, compCall)

		for paramName, resolvedValue := range resolved {
			if strings.HasPrefix(resolvedValue, "$") {
				continue
			}

			if resolvedValue != compCallName {
				continue
			}

			paramCompCalls := ast.FilterNodesInTree(compDefContent, func(n ast.Node) bool {
				if !isCompParamRefInCompDef(compDef, n) {
					return false
				}
				return getParamCompCallNameStr(n) == paramName
			})

			if len(paramCompCalls) > 0 {
				return true
			}
		}

		return false
	}
}

func isClosingParamCompCallInCycle() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return getClosingParamCompCallTargetName(ctx, node) != ""
	}
}

func getClosingParamCompCallTargetName(ctx *wrapContext, node ast.Node) string {
	return getParamCycleClosers(ctx)[node]
}

func getParamCycleClosers(ctx *wrapContext) map[ast.Node]string {
	if ctx.paramCycleClosers != nil {
		return ctx.paramCycleClosers
	}

	ctx.paramCycleClosers = map[ast.Node]string{}

	rootContent := ast.FindNodeByRuleName(ctx.root.Children(), "root-content")
	if rootContent == nil {
		return ctx.paramCycleClosers
	}

	rootCompCalls := ast.FilterNodesInTree(rootContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	for _, rootCompCall := range rootCompCalls {
		rootCompName := getCompCallNameStr(rootCompCall)
		if rootCompName == "" {
			continue
		}

		path := map[string]bool{}
		var dfs func(callNode ast.Node, compName string, resolved map[string]string)

		dfs = func(callNode ast.Node, compName string, resolved map[string]string) {
			signature := makeResolvedCallSignature(compName, resolved)
			if path[signature] {
				return
			}

			path[signature] = true
			defer delete(path, signature)

			compDef := findCompDef(ctx.root, callNode, compName)
			if compDef == nil {
				return
			}

			compDefContent := getCompDefContent(compDef)
			if compDefContent == nil {
				return
			}

			paramCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
				return isCompParamRefInCompDef(compDef, node)
			})

			for _, paramCompCall := range paramCompCalls {
				paramName := getParamCompCallNameStr(paramCompCall)
				if paramName == "" {
					continue
				}

				targetCompName := resolved[paramName]
				if targetCompName == "" || strings.HasPrefix(targetCompName, "$") {
					continue
				}

				if findCompDef(ctx.root, paramCompCall, targetCompName) == nil {
					continue
				}

				nextResolved := resolveCompArgValuesForCallTarget(ctx, paramCompCall, targetCompName, resolved)
				nextSignature := makeResolvedCallSignature(targetCompName, nextResolved)
				if path[nextSignature] {
					if _, exists := ctx.paramCycleClosers[paramCompCall]; !exists {
						ctx.paramCycleClosers[paramCompCall] = targetCompName
					}
					continue
				}

				dfs(paramCompCall, targetCompName, nextResolved)
			}
		}

		dfs(rootCompCall, rootCompName, resolveCompArgValues(ctx, rootCompCall))
	}

	return ctx.paramCycleClosers
}

func resolveCompArgValuesForCallTarget(
	ctx *wrapContext,
	callNode ast.Node,
	targetCompName string,
	parentResolvedCompArgs map[string]string,
) map[string]string {
	targetCompDef := findCompDef(ctx.root, callNode, targetCompName)
	if targetCompDef == nil {
		return nil
	}

	resolved := map[string]string{}
	for paramName, defaultValue := range getCompDefCompParamDefaults(targetCompDef) {
		resolved[paramName] = defaultValue
	}

	for _, arg := range ast.GetCompCallArgsFromCompCall(callNode) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}

		argName := ast.GetArgNameFromCompCallArg(arg)
		if argName == "" {
			continue
		}

		argType := ast.GetTypeFromCompCallArg(arg)
		switch argType {
		case "comp":
			resolved[argName] = ast.GetArgValueFromCompCallArg(arg)
		case "param":
			argValue := ast.GetArgValueFromCompCallArg(arg)
			if argValue == "" {
				continue
			}

			if parentResolvedCompArgs != nil {
				if forwardedValue, ok := parentResolvedCompArgs[argValue]; ok {
					resolved[argName] = forwardedValue
					continue
				}
			}

			resolved[argName] = "$" + argValue
		}
	}

	if len(resolved) == 0 {
		return nil
	}

	return resolved
}

func makeResolvedCallSignature(compName string, resolved map[string]string) string {
	if compName == "" {
		return ""
	}

	if len(resolved) == 0 {
		return compName
	}

	keys := make([]string, 0, len(resolved))
	for k := range resolved {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(compName)
	for _, k := range keys {
		b.WriteString("|")
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(resolved[k])
	}

	return b.String()
}

func isUnknownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		return findCompDef(ctx.root, node, compCallName) == nil
	}
}

func isKnownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !isUnknownComponent()(ctx, node)
	}
}

func hasUnknownResolvedCompArg() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		compDef := findCompDef(ctx.root, compCall, compCallName)
		if compDef == nil {
			return false
		}
		return len(getUnknownResolvedCompArgs(ctx, compCall, compDef)) > 0
	}
}

func getUnknownResolvedCompArgs(ctx *wrapContext, compCall ast.Node, compDef ast.Node) []string {
	if compDef == nil {
		return nil
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return nil
	}

	usedCompParamNames := map[string]struct{}{}
	for _, paramCompCall := range ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
		return isCompParamRefInCompDef(compDef, node)
	}) {
		name := getParamCompCallNameStr(paramCompCall)
		if name == "" {
			continue
		}
		usedCompParamNames[name] = struct{}{}
	}

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	if len(resolvedCompArgs) == 0 {
		return nil
	}

	explicitCompArgs := getExplicitCompArgMap(compCall)
	var unknowns []string

	for _, info := range getCompDefParamInfos(compDef) {
		if info.typ != "comp" {
			continue
		}
		if _, used := usedCompParamNames[info.name]; !used {
			continue
		}

		value := resolvedCompArgs[info.name]
		if value == "" || strings.HasPrefix(value, "$") {
			continue
		}

		lookupScope := compCall
		if _, ok := explicitCompArgs[info.name]; !ok && ast.IsRuleName(compDef, "global-comp-def") {
			if globalCompContent := getCompDefContent(compDef); globalCompContent != nil {
				lookupScope = globalCompContent
			}
		}

		if findCompDef(ctx.root, lookupScope, value) == nil && !util.InSliceString(value, unknowns) {
			unknowns = append(unknowns, value)
		}
	}

	return unknowns
}

func callsBlockComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		if compCallName == "" {
			return false
		}

		compDef := findCompDef(ctx.root, node, compCallName)
		if compDef == nil {
			return false
		}

		return isBlockComponent(compDef)
	}
}

func resolvedCompArgUsedAsInlineButIsBlock() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		return getResolvedInlineBlockCompName(ctx, compCall) != ""
	}
}

func getResolvedInlineBlockCompName(ctx *wrapContext, compCall ast.Node) string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return ""
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return ""
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return ""
	}

	resolved := resolveCompArgValues(ctx, compCall)
	explicitCompArgs := getExplicitCompArgMap(compCall)
	explicitParamArgs := getExplicitParamArgMap(compCall)

	inlineParamCalls := ast.FilterNodesInTree(compDefContent, func(n ast.Node) bool {
		return ast.IsRuleName(n, "param-ref") && isInlineParamRefNode(n)
	})

	for _, ipc := range inlineParamCalls {
		hasKeyAccessor := false
		for _, accessor := range ast.GetParamRefAccessors(ipc) {
			if accessor.Kind == "key" {
				hasKeyAccessor = true
				break
			}
		}
		if hasKeyAccessor {
			continue
		}

		ipcName := getParamRefNameStr(ipc)
		if ipcName == "" {
			continue
		}

		resolvedValue, ok := resolveInlineParamRefValue(ctx, compCall, ipc, resolved)
		if !ok || resolvedValue.Type != "comp" || resolvedValue.Raw == "" {
			continue
		}

		resolvedCompName := resolvedValue.Raw

		lookupScope := compCall
		_, hasExplicitCompArg := explicitCompArgs[ipcName]
		_, hasExplicitParamArg := explicitParamArgs[ipcName]
		if !hasExplicitCompArg && !hasExplicitParamArg && ast.IsRuleName(compDef, "global-comp-def") {
			if globalCompContent := getCompDefContent(compDef); globalCompContent != nil {
				lookupScope = globalCompContent
			}
		}

		argCompDef := findCompDef(ctx.root, lookupScope, resolvedCompName)
		if argCompDef == nil {
			continue
		}

		if isBlockComponent(argCompDef) {
			return resolvedCompName
		}

		if !hasExplicitCompArg || hasExplicitParamArg || !ast.IsRuleName(compDef, "global-comp-def") {
			continue
		}

		globalCompContent := getCompDefContent(compDef)
		if globalCompContent == nil || len(globalCompContent.Children()) <= 1 {
			continue
		}

		globalScopedCompDef := findCompDef(ctx.root, globalCompContent, resolvedCompName)
		if globalScopedCompDef == nil {
			return resolvedCompName
		}
	}

	return ""
}

func resolveInlineParamRefValue(
	ctx *wrapContext,
	compCall ast.Node,
	paramRef ast.Node,
	resolved map[string]string,
) (ast.ResolvedValue, bool) {
	paramName := getParamRefNameStr(paramRef)
	if paramName == "" {
		return ast.ResolvedValue{}, false
	}

	accessors := ast.GetParamRefAccessors(paramRef)
	if len(accessors) > 0 {
		if explicitArg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), paramName); explicitArg != nil {
			return ast.ApplyAccessors(
				ast.ResolveCompCallArgValue(ctx.root, explicitArg, ast.GetAncestors(compCall), compCall),
				accessors,
			), true
		}

		return ast.ApplyAccessors(
			ast.ResolveParamDefaultFromCompCall(ctx.root, compCall, paramName),
			accessors,
		), true
	}

	resolvedCompName, ok := resolved[paramName]
	if !ok || resolvedCompName == "" || strings.HasPrefix(resolvedCompName, "$") {
		return ast.ResolvedValue{}, false
	}

	return ast.ResolvedValue{
		Type:  "comp",
		Raw:   resolvedCompName,
		Scope: ast.GetLocalCompSourceFromNode(compCall, ctx.root),
	}, true
}

func isNotCompParamCompCall() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		if len(ast.GetParamRefAccessors(node)) > 0 {
			return false
		}

		paramName := getParamCompCallNameStr(node)
		if paramName == "" {
			return false
		}

		compDef := findEnclosingCompDef(node)
		if compDef == nil {
			return false
		}

		for _, info := range getCompDefParamInfos(compDef) {
			if info.name != paramName {
				continue
			}
			return info.typ != "" && info.typ != "comp"
		}

		return false
	}
}

func hasUndefinedArgs() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		return len(getUndefinedArgNames(ctx, compCall)) > 0
	}
}

func getUndefinedArgNames(ctx *wrapContext, compCall ast.Node) []string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return []string{}
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return []string{}
	}
	definedParams := getCompDefParamNames(compDef)

	undefined := make([]string, 0)
	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}

		argName := ast.GetArgNameFromCompCallArg(arg)
		if util.InSliceString(argName, definedParams) {
			continue
		}

		undefined = append(undefined, argName)
	}

	undefined = appendUniqueStrings(undefined, getUndefinedArgNamesFromResolvedParamCompCalls(ctx, compCall)...)

	return undefined
}

func hasWrongTypeArgs() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		return len(getWrongTypeArgNames(ctx, compCall)) > 0
	}
}

func getWrongTypeArgNames(ctx *wrapContext, compCall ast.Node) []string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return []string{}
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return []string{}
	}
	paramTypeMap := getCompDefParamTypeMap(compDef)

	wrongTypeArgNames := make([]string, 0)
	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}

		argNameStr := ast.GetArgNameFromCompCallArg(arg)
		expectedType, ok := paramTypeMap[argNameStr]
		if !ok {
			continue
		}

		actualType := ast.GetTypeFromCompCallArg(arg)
		if actualType == "" || actualType == "param" || actualType == expectedType {
			continue
		}

		wrongTypeArgNames = append(wrongTypeArgNames, argNameStr)
	}

	wrongTypeArgNames = appendUniqueStrings(wrongTypeArgNames, getWrongTypeArgNamesFromResolvedParamCompCalls(ctx, compCall)...)
	wrongTypeArgNames = appendUniqueStrings(wrongTypeArgNames, getWrongTypeArgNamesFromNestedCompCalls(ctx, compCall)...)

	return wrongTypeArgNames
}

func getWrongTypeArgNamesFromNestedCompCalls(ctx *wrapContext, compCall ast.Node) []string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return nil
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return nil
	}

	nestedCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	result := []string{}
	for _, nestedCompCall := range nestedCompCalls {
		targetCompDef := findCompDef(ctx.root, nestedCompCall, getCompCallNameStr(nestedCompCall))
		if targetCompDef == nil {
			continue
		}

		targetParamTypeMap := getCompDefParamTypeMap(targetCompDef)
		for _, arg := range ast.GetCompCallArgsFromCompCall(nestedCompCall) {
			if !ast.IsRuleName(arg, "comp-call-arg") {
				continue
			}

			argName := ast.GetArgNameFromCompCallArg(arg)
			expectedType, ok := targetParamTypeMap[argName]
			if !ok {
				continue
			}

			actualType := getResolvedArgTypeForNestedParamCompCall(ctx, compCall, arg)
			if actualType == "" || actualType == expectedType {
				continue
			}

			result = appendUniqueStrings(result, argName)
		}
	}

	return result
}

func getUndefinedArgNamesFromResolvedParamCompCalls(ctx *wrapContext, compCall ast.Node) []string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return nil
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return nil
	}

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	if len(resolvedCompArgs) == 0 {
		return nil
	}

	paramCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
		return isCompParamRefInCompDef(compDef, node) && hasCompCallArgsNode(node)
	})

	result := []string{}
	for _, paramCompCall := range paramCompCalls {
		targetCompName, targetCompDef := resolveParamCompCallTarget(ctx, compCall, paramCompCall, resolvedCompArgs)
		if targetCompName == "" || targetCompDef == nil {
			continue
		}

		targetParamNames := getCompDefParamNames(targetCompDef)

		for _, arg := range ast.GetCompCallArgsFromCompCall(paramCompCall) {
			if !ast.IsRuleName(arg, "comp-call-arg") {
				continue
			}

			argName := ast.GetArgNameFromCompCallArg(arg)
			if util.InSliceString(argName, targetParamNames) {
				continue
			}

			result = appendUniqueStrings(result, argName)
		}
	}

	return result
}

func getWrongTypeArgNamesFromResolvedParamCompCalls(ctx *wrapContext, compCall ast.Node) []string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return nil
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return nil
	}

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	if len(resolvedCompArgs) == 0 {
		return nil
	}

	paramCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
		return isCompParamRefInCompDef(compDef, node) && hasCompCallArgsNode(node)
	})

	result := []string{}
	for _, paramCompCall := range paramCompCalls {
		targetCompName, targetCompDef := resolveParamCompCallTarget(ctx, compCall, paramCompCall, resolvedCompArgs)
		if targetCompName == "" || targetCompDef == nil {
			continue
		}

		targetParamTypeMap := getCompDefParamTypeMap(targetCompDef)

		for _, arg := range ast.GetCompCallArgsFromCompCall(paramCompCall) {
			if !ast.IsRuleName(arg, "comp-call-arg") {
				continue
			}

			argName := ast.GetArgNameFromCompCallArg(arg)
			expectedType, ok := targetParamTypeMap[argName]
			if !ok {
				continue
			}

			actualType := getResolvedArgTypeForNestedParamCompCall(ctx, compCall, arg)
			if actualType == "" || actualType == expectedType {
				continue
			}

			result = appendUniqueStrings(result, argName)
		}
	}

	return result
}

func getResolvedArgTypeForNestedParamCompCall(ctx *wrapContext, compCall ast.Node, arg ast.Node) string {
	actualType := ast.GetTypeFromCompCallArg(arg)
	if actualType != "param" {
		return actualType
	}

	raw := ast.GetArgValueFromCompCallArg(arg)
	paramName, accessors := ast.GetValuePathFromRaw(raw)
	if paramName == "" {
		return ""
	}

	explicitArg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), paramName)
	if explicitArg != nil {
		return ast.ApplyAccessors(
			ast.ResolveCompCallArgValue(ctx.root, explicitArg, ast.GetAncestors(compCall), compCall),
			accessors,
		).Type
	}

	return ast.ApplyAccessors(
		ast.ResolveParamDefaultFromCompCall(ctx.root, compCall, paramName),
		accessors,
	).Type
}

func resolveParamCompCallTarget(
	ctx *wrapContext,
	compCall ast.Node,
	paramCompCall ast.Node,
	resolvedCompArgs map[string]string,
) (string, ast.Node) {
	paramName := getParamCompCallNameStr(paramCompCall)
	if paramName == "" {
		return "", nil
	}

	targetCompName, ok := resolvedCompArgs[paramName]
	if !ok || targetCompName == "" || strings.HasPrefix(targetCompName, "$") {
		return "", nil
	}

	targetCompDef := findCompDef(ctx.root, compCall, targetCompName)
	if targetCompDef == nil {
		return "", nil
	}

	return targetCompName, targetCompDef
}

func appendUniqueStrings(dst []string, src ...string) []string {
	for _, item := range src {
		if item == "" || util.InSliceString(item, dst) {
			continue
		}
		dst = append(dst, item)
	}
	return dst
}

func isInsideRootContent() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		ancestors := ast.GetAncestors(node)
		compDef := ast.FindNode(ancestors, func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
		})
		return compDef == nil
	}
}

func isUndefinedParamRef() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, paramRef ast.Node) bool {
		refName := getParamRefNameStr(paramRef)
		if refName == "" {
			return false
		}

		ancestors := ast.GetAncestors(paramRef)
		globalCompDef := ast.FindNode(ancestors, func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "global-comp-def")
		})
		localCompDef := ast.FindNode(ancestors, func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "local-comp-def")
		})

		if globalCompDef == nil && localCompDef == nil {
			return false
		}

		if globalCompDef != nil && localCompDef == nil {
			return !util.InSliceString(refName, getCompDefParamNames(globalCompDef))
		}

		if globalCompDef == nil && localCompDef != nil {
			return !util.InSliceString(refName, getCompDefParamNames(localCompDef))
		}

		if util.InSliceString(refName, getCompDefParamNames(localCompDef)) {
			return false
		}

		return !util.InSliceString(refName, getCompDefParamNames(globalCompDef))
	}
}

func isUndefinedParamCompCall() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		paramName := getParamCompCallNameStr(node)
		if paramName == "" {
			return false
		}

		compDef := findEnclosingCompDef(node)
		if compDef == nil {
			return true
		}

		definedParams := getCompDefParamNames(compDef)

		return !util.InSliceString(paramName, definedParams)
	}
}

func resolveCompArgValues(ctx *wrapContext, compCall ast.Node) map[string]string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return nil
	}

	defCompParams := getCompDefCompParamDefaults(compDef)

	explicitCompArgs := getExplicitCompArgMap(compCall)

	explicitParamArgs := getExplicitParamArgMap(compCall)

	result := make(map[string]string)
	for paramName, defaultValue := range defCompParams {
		if explicit, ok := explicitCompArgs[paramName]; ok {
			result[paramName] = explicit
		} else if paramRef, ok := explicitParamArgs[paramName]; ok {
			result[paramName] = "$" + paramRef
		} else {
			result[paramName] = defaultValue
		}
	}

	return result
}

func getExplicitCompArgMap(compCall ast.Node) map[string]string {
	return getExplicitArgMapByType(compCall, "comp")
}

func getExplicitParamArgMap(compCall ast.Node) map[string]string {
	return getExplicitArgMapByType(compCall, "param")
}

func getExplicitArgMapByType(compCall ast.Node, typ string) map[string]string {
	result := make(map[string]string)

	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}
		if ast.GetTypeFromCompCallArg(arg) != typ {
			continue
		}
		result[ast.GetArgNameFromCompCallArg(arg)] = ast.GetArgValueFromCompCallArg(arg)
	}

	return result
}

func getCompDefCompParamDefaults(compDef ast.Node) map[string]string {
	result := map[string]string{}
	for _, info := range getCompDefParamInfos(compDef) {
		if info.typ == "comp" && info.defVal != "" {
			result[info.name] = info.defVal
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func getCompCallNameStr(node ast.Node) string {
	compCallNameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
	if compCallNameNode != nil {
		return strings.TrimSpace(string(compCallNameNode.Raw()))
	}
	return ""
}

func getParamRefNameStr(node ast.Node) string {
	return ast.GetParamRefName(node)
}

func getParamCompCallNameStr(node ast.Node) string {
	nameNode := ast.FindNodeByRuleName(node.Children(), "param-comp-call-name")
	if nameNode != nil {
		return strings.TrimSpace(string(nameNode.Raw()))
	}
	return getParamRefNameStr(node)
}

func getCompDefContent(compDef ast.Node) ast.Node {
	for _, child := range compDef.Children() {
		if child.Rule() == nil {
			continue
		}
		ruleName := child.Rule().Name()
		if ruleName == "local-comp-def-content" || ruleName == "global-comp-def-content" {
			return child
		}
	}
	return nil
}

func getCompDefParamInfos(compDef ast.Node) []compParamInfo {
	compParams := ast.GetCompParamsFromCompDef(compDef)
	if len(compParams) == 0 {
		return nil
	}

	result := make([]compParamInfo, 0, len(compParams))
	for _, compParam := range compParams {
		if !ast.IsRuleName(compParam, "comp-param") {
			continue
		}

		name := ast.GetParamNameFromCompParam(compParam)
		if name == "" {
			continue
		}

		defVal := ""
		compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
		if compParamType != nil && len(compParamType.Children()) > 0 {
			typeVariant := compParamType.Children()[0]
			if ast.FindNodeByRuleName(typeVariant.Children(), "comp-param-defa-value") != nil {
				defVal = ast.GetParamDefValFromCompParam(compParam)
			}
		}

		result = append(result, compParamInfo{
			name:   name,
			typ:    ast.GetTypeFromCompParam(compParam),
			defVal: defVal,
		})
	}

	return result
}

func getCompDefParamNames(compDef ast.Node) []string {
	names := make([]string, 0)
	for _, info := range getCompDefParamInfos(compDef) {
		names = append(names, info.name)
	}
	return names
}

func getCompDefParamTypeMap(compDef ast.Node) map[string]string {
	result := map[string]string{}
	for _, info := range getCompDefParamInfos(compDef) {
		result[info.name] = info.typ
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func findEnclosingCompDef(node ast.Node) ast.Node {
	ancestors := ast.GetAncestors(node)
	return ast.FindNode(ancestors, func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
	})
}

func isInlineParamRefNode(node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}

	if ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "p-content")
	}) != nil {
		pContent := ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "p-content")
		})
		if pContent == nil {
			return false
		}

		hasSoftBreak := false
		for _, child := range pContent.Children() {
			if ast.IsRuleName(child, "soft-break") {
				hasSoftBreak = true
			}
		}
		if hasSoftBreak {
			return false
		}

		for _, child := range pContent.Children() {
			if child == node {
				continue
			}
			if ast.IsRuleName(child, "plain") && strings.TrimSpace(string(child.Raw())) == "" {
				continue
			}
			return true
		}
		return false
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

func isBlockComponent(compDef ast.Node) bool {
	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return false
	}

	childrenCount := len(compDefContent.Children())
	if childrenCount == 0 {
		return false
	} else if childrenCount > 1 {
		return true
	}

	p := ast.FindNodeByRuleName(compDefContent.Children(), "p")
	if p == nil {
		return true
	}

	pContent := ast.FindNodeByRuleName(p.Children(), "p-content")
	softBlock := ast.FindNodeByRuleName(pContent.Children(), "soft-break")

	return softBlock != nil
}

func findCompDef(root ast.Node, compCallNode ast.Node, name string) ast.Node {
	globalCompDefAnc := ast.FindNode(ast.GetAncestors(compCallNode), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := ast.FindLocalCompDef(localCompDefSrc, name)
	if localCompDef != nil {
		return localCompDef
	}

	globalCompDef := ast.FindGlobalCompDef(root, name)
	if globalCompDef != nil {
		return globalCompDef
	}

	builtinCompDef := ast.FindBuiltinCompDef(root, name)
	if builtinCompDef != nil {
		return builtinCompDef
	}

	return nil
}
