package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/util"
)

func invalidBuiltinCompCallSchema() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			not(func(_ *wrapContext, node ast.Node) bool { return findEnclosingCompDef(node) != nil }),
			isKnownComponent(),
			hasBuiltinSchemaMismatches(),
		},
		title:   invalidBuiltinCompCallSchemaTitle,
		message: invalidBuiltinCompCallSchemaMsg,
		block:   blockFromRuleName,
	}
}

func undefinedParam() conditionAnalyzer {
	return conditionAnalyzer{
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

func wrongArgType() conditionAnalyzer {
	return conditionAnalyzer{
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
	mismatches := getBuiltinSchemaMismatchesForCompCall(ctx, ownerCompCall, targetCompCall)
	names := make([]string, 0, len(mismatches))
	for _, mismatch := range mismatches {
		names = appendUniqueStrings(names, mismatch.name)
	}
	return names
}

type builtinSchemaMismatch struct {
	name       string
	diagnostic builtin.ValidationDiagnostic
}

func getBuiltinSchemaMismatchesForCompCall(ctx *wrapContext, ownerCompCall ast.Node, targetCompCall ast.Node) []builtinSchemaMismatch {
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

	paramByName := make(map[string]builtin.Param, len(definition.Params))
	for _, param := range definition.Params {
		paramByName[param.Name] = param
	}

	invokerAncestors := ast.GetAncestors(ownerCompCall)
	currentCompCall := ownerCompCall
	if ownerCompCall != targetCompCall {
		invokerAncestors = append([]ast.Node{targetCompCall, ownerCompCall}, invokerAncestors...)
		currentCompCall = targetCompCall
	}

	mismatches := []builtinSchemaMismatch{}
	seenArgs := map[string]bool{}
	for _, arg := range ast.GetCompCallArgsFromCompCall(targetCompCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}

		argName := ast.GetArgNameFromCompCallArg(arg)
		param, ok := paramByName[argName]
		if !ok {
			continue
		}
		seenArgs[argName] = true

		resolved := ast.ResolveCompCallArgValue(ctx.root, arg, invokerAncestors, currentCompCall)
		if resolved.IsZero() || resolvedValueMissingContextKey(resolved) != "" || builtin.MatchesResolvedValue(param.Schema, resolved) {
			continue
		}

		mismatches = appendBuiltinSchemaMismatch(mismatches, builtinSchemaMismatch{
			name:       argName,
			diagnostic: builtinParamDiagnostic(param, resolved),
		})
	}

	for _, param := range definition.Params {
		if !param.IsRequired || seenArgs[param.Name] {
			continue
		}

		resolved := ast.ResolveParamDefaultFromCompCall(ctx.root, targetCompCall, param.Name)
		if resolved.IsZero() || resolvedValueMissingContextKey(resolved) != "" || builtin.MatchesResolvedValue(param.Schema, resolved) {
			continue
		}

		mismatches = appendBuiltinSchemaMismatch(mismatches, builtinSchemaMismatch{
			name:       param.Name,
			diagnostic: builtinParamDiagnostic(param, resolved),
		})
	}

	return mismatches
}

func appendBuiltinSchemaMismatch(values []builtinSchemaMismatch, next builtinSchemaMismatch) []builtinSchemaMismatch {
	for _, value := range values {
		if value.name == next.name {
			return values
		}
	}
	return append(values, next)
}

func builtinParamDiagnostic(param builtin.Param, value ast.ResolvedValue) builtin.ValidationDiagnostic {
	if param.Diagnostic == nil {
		return builtin.ValidationDiagnostic{}
	}
	diagnostic, ok := param.Diagnostic(param.Name, value)
	if !ok {
		return builtin.ValidationDiagnostic{}
	}
	return diagnostic
}

func getBuiltinSchemaMismatchDiagnostic(ctx *wrapContext, compCall ast.Node) builtin.ValidationDiagnostic {
	if diagnostic := firstBuiltinSchemaMismatchDiagnostic(getBuiltinSchemaMismatchesForCompCall(ctx, compCall, compCall)); diagnostic.Title != "" {
		return diagnostic
	}

	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return builtin.ValidationDiagnostic{}
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return builtin.ValidationDiagnostic{}
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return builtin.ValidationDiagnostic{}
	}

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	if len(resolvedCompArgs) > 0 {
		paramCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
			return isCompParamRefInCompDef(compDef, node) && hasCompCallArgsNode(node)
		})

		for _, paramCompCall := range paramCompCalls {
			_, targetCompDef := resolveParamCompCallTarget(ctx, compCall, paramCompCall, resolvedCompArgs)
			if targetCompDef == nil || !ast.IsRuleName(targetCompDef, "builtin-comp") {
				continue
			}
			if diagnostic := firstBuiltinSchemaMismatchDiagnostic(getBuiltinSchemaMismatchesForCompCall(ctx, compCall, paramCompCall)); diagnostic.Title != "" {
				return diagnostic
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
		if diagnostic := firstBuiltinSchemaMismatchDiagnostic(getBuiltinSchemaMismatchesForCompCall(ctx, compCall, nestedCompCall)); diagnostic.Title != "" {
			return diagnostic
		}
	}

	return builtin.ValidationDiagnostic{}
}

func firstBuiltinSchemaMismatchDiagnostic(mismatches []builtinSchemaMismatch) builtin.ValidationDiagnostic {
	for _, mismatch := range mismatches {
		if mismatch.diagnostic.Title != "" {
			return mismatch.diagnostic
		}
	}
	return builtin.ValidationDiagnostic{}
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
