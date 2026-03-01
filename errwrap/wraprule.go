package errwrap

import (
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
	root           ast.Node
	compCallChains [][]ast.Node
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
		unknownCompCall(),
		unknownCompParamCall(),
		blockCompInsideInline(),
		blockParamCompInsideInline(),
		undefinedParam(),
		wrongArgType(),
		paramRefInRootContent(),
		paramCompCallInRootContent(),
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
		message: wrongArgTypeMsgFn,
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

func paramCompCallInRootContent() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-param-comp-call", "inline-param-comp-call"),
			isInsideRootContent(),
		},
		title:   staticTitle("Parameter component call in root content"),
		message: paramCompCallInRootMsg,
		block:   blockFromRuleName,
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
		block:   blockFromRuleName,
	}
}

func undefinedParamCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-param-comp-call", "inline-param-comp-call"),
			not(isInsideRootContent()),
			isUndefinedParamCompCall(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamCompCallAsUnknownMsg,
		block:   blockFromRuleName,
	}
}

func notCompParamCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-param-comp-call", "inline-param-comp-call"),
			not(isInsideRootContent()),
			isNotCompParamCompCall(),
		},
		title:   staticTitle("Not component parameter"),
		message: notCompParamCompCallMsg,
		block:   blockFromRuleName,
	}
}

func alwaysBlock(_ *wrapContext, _ ast.Node) bool { return true }
func neverBlock(_ *wrapContext, _ ast.Node) bool  { return false }
func blockFromRuleName(_ *wrapContext, node ast.Node) bool {
	return strings.HasPrefix(node.Rule().Name(), "block-")
}

func staticTitle(s string) func(*wrapContext, ast.Node) string {
	return func(_ *wrapContext, _ ast.Node) string { return s }
}

func infiniteCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
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

func wrongArgTypeMsgFn(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "One or more arguments passed to **" + name + "** have the wrong type. Make sure each argument matches the expected type from the component definition."
}

func paramRefInRootMsg(_ *wrapContext, _ ast.Node) string {
	return "Parameters cannot be used in the root context."
}

func paramCompCallInRootMsg(_ *wrapContext, node ast.Node) string {
	name := getParamCompCallNameStr(node)
	return "The parameter component call **{{ $" + name + " }}** cannot be used outside of a component definition."
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

func not(cond func(*wrapContext, ast.Node) bool) func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !cond(ctx, node)
	}
}

func isCalledByItself() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		compCallName := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
		if compCallName == nil {
			return false
		}
		compCallNameStr := strings.TrimSpace(string(compCallName.Raw()))

		compDefs := ast.FilterNodes(ast.GetAncestors(node), func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
		})

		if len(compDefs) == 0 {
			return false
		}

		called := ast.FindNode(compDefs, func(def ast.Node) bool {
			return getCompDefName(def) == compCallNameStr
		})

		return called != nil
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
					return true
				}
			}
		}

		return false
	}
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
				if !ast.IsRuleNameOneOf(n, []string{"block-param-comp-call", "inline-param-comp-call"}) {
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

func isUnknownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		if isBuiltInComp(compCallName) {
			return false
		}
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

		value := resolvedCompArgs[info.name]
		if value == "" || strings.HasPrefix(value, "$") || isBuiltInComp(value) {
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

	inlineParamCalls := ast.FilterNodesInTree(compDefContent, func(n ast.Node) bool {
		return ast.IsRuleName(n, "inline-param-comp-call")
	})

	for _, ipc := range inlineParamCalls {
		ipcName := getParamCompCallNameStr(ipc)
		if ipcName == "" {
			continue
		}

		resolvedCompName, ok := resolved[ipcName]
		if !ok || resolvedCompName == "" {
			continue
		}

		if strings.HasPrefix(resolvedCompName, "$") {
			continue
		}

		argCompDef := findCompDef(ctx.root, compCall, resolvedCompName)
		if argCompDef == nil {
			continue
		}

		if isBlockComponent(argCompDef) {
			return resolvedCompName
		}
	}

	return ""
}

func isNotCompParamCompCall() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
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
	if isBuiltInComp(compCallName) {
		definedParams = getBuiltInCompParams(compCallName)
	}

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

	return undefined
}

func hasWrongTypeArgs() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		if compCallName == "" {
			return false
		}

		compDef := findCompDef(ctx.root, compCall, compCallName)
		if compDef == nil {
			return false
		}

		paramTypeMap := getCompDefParamTypeMap(compDef)
		if isBuiltInComp(compCallName) {
			paramTypeMap = getBuiltInCompParamTypes(compCallName)
		}

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
			if actualType == "" {
				continue
			}

			if actualType == "param" {
				continue
			}

			if actualType != expectedType {
				return true
			}
		}

		return false
	}
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

func getCompDefName(def ast.Node) string {
	if ast.IsRuleName(def, "local-comp-def") {
		head := ast.FindNodeByRuleName(def.Children(), "local-comp-def-head")
		if head == nil {
			return ""
		}
		compName := ast.FindNodeByRuleName(head.Children(), "local-comp-name")
		if compName == nil {
			return ""
		}
		return strings.TrimSpace(string(compName.Raw()))
	}
	if ast.IsRuleName(def, "global-comp-def") {
		compName := ast.FindNodeByRuleName(def.Children(), "global-comp-name")
		if compName == nil {
			return ""
		}
		return strings.TrimSpace(string(compName.Raw()))
	}
	return ""
}

func getCompCallNameStr(node ast.Node) string {
	compCallNameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
	if compCallNameNode != nil {
		return strings.TrimSpace(string(compCallNameNode.Raw()))
	}
	return ""
}

func getParamRefNameStr(node ast.Node) string {
	refNameNode := ast.FindNodeByRuleName(node.Children(), "param-ref-name")
	if refNameNode != nil {
		return strings.TrimSpace(string(refNameNode.Raw()))
	}
	return ""
}

func getParamCompCallNameStr(node ast.Node) string {
	nameNode := ast.FindNodeByRuleName(node.Children(), "param-comp-call-name")
	if nameNode != nil {
		return strings.TrimSpace(string(nameNode.Raw()))
	}
	return ""
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

		nameNode := ast.FindNodeByRuleName(compParam.Children(), "comp-param-name")
		if nameNode == nil {
			continue
		}

		typeNode := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
		defVal := ""
		typ := ""
		if typeNode != nil && len(typeNode.Children()) > 0 {
			typ = ast.GetTypeFromCompParam(compParam)
			typeVariant := typeNode.Children()[0]
			defNode := ast.FindNodeByRuleName(typeVariant.Children(), "comp-param-defa-value")
			if defNode != nil {
				defVal = strings.TrimSpace(string(defNode.Raw()))
			}
		}

		result = append(result, compParamInfo{
			name:   strings.TrimSpace(string(nameNode.Raw())),
			typ:    typ,
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

	if softBlock != nil {
		return true
	}

	return false
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

	return nil
}

func isBuiltInComp(name string) bool {
	return util.InSliceString(name, []string{"LINK"})
}

func getBuiltInCompParams(name string) []string {
	switch name {
	case "LINK":
		return []string{"url", "text", "new-tab"}
	}
	return nil
}

func getBuiltInCompParamTypes(name string) map[string]string {
	switch name {
	case "LINK":
		return map[string]string{
			"url":     "string",
			"text":    "string",
			"new-tab": "bool",
		}
	}
	return nil
}
