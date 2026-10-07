package errwrap

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/util"
)

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

func missingArg() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasMissingArgs(),
		},
		title:   staticTitle("Missing argument"),
		message: missingArgMsg,
		block:   blockFromRuleName,
	}
}

func wrongArgType() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			isNotBuiltinComponent(),
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

	compDef := ast.FindCompDef(ctx.root, compCall, compCallName)
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

	undefined = appendUniqueStrings(undefined, getUndefinedParamArgRefNames(compCall)...)
	undefined = appendUniqueStrings(undefined, getUndefinedBoundArgNames(ctx.root, compCall)...)

	return undefined
}

func hasMissingArgs() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		_, missing := getMissingArgs(ctx, compCall)
		return len(missing) > 0
	}
}

// getMissingArgs returns the component whose required parameters are missing and
// the missing parameter names in definition order. Required parameters of the
// called definition are reduced by the arguments given by the caller and, for a
// bound component value, by its bound arguments. Built-in components are never
// checked.
func getMissingArgs(ctx *wrapContext, compCall ast.Node) (string, []string) {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return "", nil
	}

	compDef := ast.FindCompDef(ctx.root, compCall, compCallName)
	if compDef == nil || ast.IsRuleName(compDef, "builtin-comp") {
		return "", nil
	}

	missing := missingArgNames(getRequiredParamNames(compDef), getCallArgNames(compCall))
	if len(missing) > 0 {
		return compCallName, missing
	}

	return "", nil
}

func getRequiredParamNames(compDef ast.Node) []string {
	names := []string{}
	for _, compParam := range ast.GetCompParamsFromCompDef(compDef) {
		if !ast.IsRuleName(compParam, "comp-param") || !ast.IsRequiredCompParam(compParam) {
			continue
		}

		name := ast.GetParamNameFromCompParam(compParam)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}

func getCallArgNames(compCall ast.Node) []string {
	names := []string{}
	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}
		names = appendUniqueStrings(names, ast.GetArgNameFromCompCallArg(arg))
	}
	return names
}

func missingArgNames(required []string, supplied []string) []string {
	missing := []string{}
	for _, name := range required {
		if util.InSliceString(name, supplied) {
			continue
		}
		missing = appendUniqueStrings(missing, name)
	}
	return missing
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

	compDef := ast.FindCompDef(ctx.root, compCall, compCallName)
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

	wrongTypeArgNames = appendUniqueStrings(wrongTypeArgNames, getWrongTypeBoundArgNames(ctx.root, compCall)...)

	return wrongTypeArgNames
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

type builtinSchemaMismatch struct {
	name       string
	diagnostic builtin.ValidationDiagnostic
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

func firstBuiltinSchemaMismatchDiagnostic(mismatches []builtinSchemaMismatch) builtin.ValidationDiagnostic {
	for _, mismatch := range mismatches {
		if mismatch.diagnostic.Title != "" {
			return mismatch.diagnostic
		}
	}
	return builtin.ValidationDiagnostic{}
}
