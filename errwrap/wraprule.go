package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/util"
)

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

func diagnosticAnalyzers() []diagnosticAnalyzer {
	return []diagnosticAnalyzer{
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
		wrongImageArgType(),
		invalidWebGrid(),
		invalidImage(),
		invalidBuiltinCompCallSchema(),
		unknownWebGridItemComponent(),
		wrongArgType(),
		paramRefInRootContent(),
		contextRefAnalyzer{},
		undefinedParamRef(),
		notCompParamCompCall(),
		undefinedParamCompCall(),
	}
}

func invalidBuiltinCompCallSchema() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			not(func(_ *wrapContext, node ast.Node) bool { return findEnclosingCompDef(node) != nil }),
			isKnownComponent(),
			hasBuiltinSchemaMismatches(),
		},
		title:   staticTitle("Invalid built-in arguments"),
		message: invalidBuiltinCompCallSchemaMsg,
		block:   blockFromRuleName,
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

func hasBuiltinSchemaMismatches() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		return len(getBuiltinSchemaMismatchArgNames(ctx, compCall)) > 0
	}
}

func getBuiltinSchemaMismatchArgNames(ctx *wrapContext, compCall ast.Node) []string {
	result := getBuiltinSchemaMismatchArgNamesForCompCall(ctx, compCall, compCall)
	result = appendUniqueStrings(result, getBuiltinSchemaMismatchArgNamesFromResolvedParamCompCalls(ctx, compCall)...)
	result = appendUniqueStrings(result, getBuiltinSchemaMismatchArgNamesFromNestedCompCalls(ctx, compCall)...)
	return result
}

func getBuiltinSchemaMismatchTargetName(ctx *wrapContext, compCall ast.Node) string {
	if len(getBuiltinSchemaMismatchArgNamesForCompCall(ctx, compCall, compCall)) > 0 {
		return getCompCallNameStr(compCall)
	}

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

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	if len(resolvedCompArgs) > 0 {
		paramCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
			return isCompParamRefInCompDef(compDef, node) && hasCompCallArgsNode(node)
		})

		for _, paramCompCall := range paramCompCalls {
			targetCompName, targetCompDef := resolveParamCompCallTarget(ctx, compCall, paramCompCall, resolvedCompArgs)
			if targetCompName == "" || targetCompDef == nil || !ast.IsRuleName(targetCompDef, "builtin-comp") {
				continue
			}
			if len(getBuiltinSchemaMismatchArgNamesForCompCall(ctx, compCall, paramCompCall)) > 0 {
				return targetCompName
			}
		}
	}

	nestedCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	for _, nestedCompCall := range nestedCompCalls {
		nestedName := getCompCallNameStr(nestedCompCall)
		nestedDef := findCompDef(ctx.root, nestedCompCall, nestedName)
		if nestedName == "" || nestedDef == nil || !ast.IsRuleName(nestedDef, "builtin-comp") {
			continue
		}
		if len(getBuiltinSchemaMismatchArgNamesForCompCall(ctx, compCall, nestedCompCall)) > 0 {
			return nestedName
		}
	}

	return ""
}

func getBuiltinSchemaMismatchArgNamesForCompCall(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node) []string {
	compName := getCompCallNameStr(targetCompCall)
	if compName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, targetCompCall, compName)
	if compDef == nil || !ast.IsRuleName(compDef, "builtin-comp") {
		return nil
	}

	definition, ok := builtin.FindDefinition(compName)
	if !ok {
		return nil
	}

	paramSchemaByName := make(map[string]builtin.ValueSchema, len(definition.Params))
	for _, param := range definition.Params {
		paramSchemaByName[param.Name] = param.Schema
	}

	mismatches := []string{}
	for _, arg := range ast.GetCompCallArgsFromCompCall(targetCompCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}

		argName := ast.GetArgNameFromCompCallArg(arg)
		schema, ok := paramSchemaByName[argName]
		if !ok {
			continue
		}

		invokerAncestors := ast.GetAncestors(ownerCompCall)
		currentCompCall := ownerCompCall
		if ownerCompCall != targetCompCall {
			invokerAncestors = append([]ast.Node{targetCompCall, ownerCompCall}, invokerAncestors...)
			currentCompCall = targetCompCall
		}

		resolved := ast.ResolveCompCallArgValue(ctx.root, arg, invokerAncestors, currentCompCall)
		if resolved.IsZero() || builtin.MatchesResolvedValue(schema, resolved) {
			continue
		}

		mismatches = appendUniqueStrings(mismatches, argName)
	}

	return mismatches
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
		if !ok || expectedType == "" {
			continue
		}

		actualType := ast.GetTypeFromCompCallArg(arg)
		if actualType == "context" {
			actualType = ast.ResolveCompCallArgValue(ctx.root, arg, ast.GetAncestors(compCall), compCall).Type
		}
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
			if !ok || expectedType == "" {
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

func getBuiltinSchemaMismatchArgNamesFromNestedCompCalls(ctx *wrapContext, compCall ast.Node) []string {
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
		result = appendUniqueStrings(result, getBuiltinSchemaMismatchArgNamesForCompCall(ctx, compCall, nestedCompCall)...)
	}

	return result
}

func getUndefinedArgNamesFromResolvedParamCompCalls(ctx *wrapContext, compCall ast.Node) []string {
	return collectFromResolvedParamCompCalls(ctx, compCall, func(paramCompCall ast.Node, _ string, targetCompDef ast.Node) []string {
		targetParamNames := getCompDefParamNames(targetCompDef)

		result := []string{}
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

		return result
	})
}

func getWrongTypeArgNamesFromResolvedParamCompCalls(ctx *wrapContext, compCall ast.Node) []string {
	return collectFromResolvedParamCompCalls(ctx, compCall, func(paramCompCall ast.Node, _ string, targetCompDef ast.Node) []string {
		targetParamTypeMap := getCompDefParamTypeMap(targetCompDef)

		result := []string{}
		for _, arg := range ast.GetCompCallArgsFromCompCall(paramCompCall) {
			if !ast.IsRuleName(arg, "comp-call-arg") {
				continue
			}

			argName := ast.GetArgNameFromCompCallArg(arg)
			expectedType, ok := targetParamTypeMap[argName]
			if !ok || expectedType == "" {
				continue
			}

			actualType := getResolvedArgTypeForNestedParamCompCall(ctx, compCall, arg)
			if actualType == "" || actualType == expectedType {
				continue
			}

			result = appendUniqueStrings(result, argName)
		}

		return result
	})
}

func getBuiltinSchemaMismatchArgNamesFromResolvedParamCompCalls(ctx *wrapContext, compCall ast.Node) []string {
	return collectFromResolvedParamCompCalls(ctx, compCall, func(paramCompCall ast.Node, _ string, targetCompDef ast.Node) []string {
		if !ast.IsRuleName(targetCompDef, "builtin-comp") {
			return nil
		}

		return getBuiltinSchemaMismatchArgNamesForCompCall(ctx, compCall, paramCompCall)
	})
}

func collectFromResolvedParamCompCalls(
	ctx *wrapContext,
	compCall ast.Node,
	collect func(paramCompCall ast.Node, targetCompName string, targetCompDef ast.Node) []string,
) []string {
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

		result = appendUniqueStrings(result, collect(paramCompCall, targetCompName, targetCompDef)...)
	}

	return result
}

func getResolvedArgTypeForNestedParamCompCall(ctx *wrapContext, compCall ast.Node, arg ast.Node) string {
	actualType := ast.GetTypeFromCompCallArg(arg)
	if actualType == "context" {
		return ast.ResolveCompCallArgValue(ctx.root, arg, ast.GetAncestors(compCall), compCall).Type
	}
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
