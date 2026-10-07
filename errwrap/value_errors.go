package errwrap

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

// ParamRefValueError returns the error of a parameter reference unit whose
// parameter resolved to value. accessors are the accessors of the reference.
func ParamRefValueError(paramName string, value ast.ResolvedValue, accessors []ast.ValueAccessor) (title, message string, ok bool) {
	applied, accessErr := ast.ApplyAccessorsDetailed(value, accessors)

	if applied.MissingContextKey != "" {
		return "Unknown key", unknownContextKeyMsg(applied.MissingContextKey), true
	}

	switch accessErr.Kind {
	case ast.AccessErrorArrayIndexOutOfRange:
		return "Array index out of range", paramArrayIndexOutOfRangeMsg(paramName), true
	case ast.AccessErrorInvalidIndexAccess:
		return "Invalid parameter access", invalidParamIndexAccessMsg(paramName), true
	case ast.AccessErrorInvalidKeyAccess:
		return "Invalid parameter access", invalidParamKeyAccessMsg(paramName), true
	case ast.AccessErrorUnknownRecordKey:
		return "Unknown record key", unknownRecordKeyMsg(accessErr.Key), true
	}

	if len(accessors) == 0 {
		switch applied.Type {
		case "array":
			return "Invalid parameter usage", directParamArrayUsageMsg(paramName), true
		case "record":
			return "Invalid parameter usage", directParamRecordUsageMsg(paramName), true
		}
	}

	return "", "", false
}

// InlineCompValueError returns the error of a parameter reference unit that
// renders the component compDef inline.
func InlineCompValueError(compName string, compDef ast.Node) (title, message string, ok bool) {
	if isBlockComponent(compDef) {
		return "Invalid component usage", "The component **" + compName + "** is a block component and cannot be used inline.", true
	}
	return "", "", false
}

// CallArgsError returns the error of a component call whose arguments,
// including values inside arrays, records and bound arguments, use a context
// key that is not injected. invokerAncestors are as for
// ast.ResolveCompCallArgValue.
func CallArgsError(root, compCall ast.Node, invokerAncestors []ast.Node) (title, message string, ok bool) {
	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		if key := resolvedValueMissingContextKey(ast.ResolveCompCallArgValue(root, arg, invokerAncestors, compCall)); key != "" {
			return "Unknown key", unknownContextKeyMsg(key), true
		}
		if key := getMissingContextKeyInBoundArgs(root, arg); key != "" {
			return "Unknown key", unknownContextKeyMsg(key), true
		}
	}
	return "", "", false
}

// ParamCompCallError returns the error of a call through a component
// parameter, paramRef, that renders the component value. compDef is the local
// or global definition the value resolved to, nil if there is none.
// invokerAncestors are as for ast.ResolveCompCallArgValue.
func ParamCompCallError(root, paramRef ast.Node, value ast.ResolvedValue, compDef ast.Node, invokerAncestors []ast.Node) (title, message string, ok bool) {
	if compDef == nil {
		return "Unknown component", "The component **" + value.Raw + "** is not defined or not registered.", true
	}

	explicitArgs := []ast.Node{}
	for _, arg := range ast.GetCompCallArgsFromCompCall(paramRef) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}
		explicitArgs = append(explicitArgs, arg)
	}

	var boundArgs []ast.Node
	if value.Bound != nil {
		boundArgs = value.Bound.Args()
	}

	definedParams := getCompDefParamNames(compDef)

	undefined := []string{}
	for _, arg := range explicitArgs {
		name := ast.GetArgNameFromCompCallArg(arg)
		if !util.InSliceString(name, definedParams) {
			undefined = appendUniqueStrings(undefined, name)
		}
	}
	for _, arg := range boundArgs {
		name := ast.GetArgNameFromCompCallArg(arg)
		if !util.InSliceString(name, definedParams) {
			undefined = appendUniqueStrings(undefined, name)
		}
	}
	if len(undefined) > 0 {
		return "Unknown parameter", undefinedParamNamesMsg(undefined), true
	}

	supplied := []string{}
	for _, arg := range explicitArgs {
		supplied = appendUniqueStrings(supplied, ast.GetArgNameFromCompCallArg(arg))
	}
	for _, arg := range boundArgs {
		supplied = appendUniqueStrings(supplied, ast.GetArgNameFromCompCallArg(arg))
	}

	missing := missingArgNames(getRequiredParamNames(compDef), supplied)
	if len(missing) > 0 {
		return "Missing argument", missingArgNamesMsg(value.Raw, missing), true
	}

	paramTypeMap := getCompDefParamTypeMap(compDef)
	wrongType := []string{}
	for _, arg := range explicitArgs {
		argName := ast.GetArgNameFromCompCallArg(arg)
		expectedType, ok := paramTypeMap[argName]
		if !ok || expectedType == "" {
			continue
		}

		actualType := ast.GetTypeFromCompCallArg(arg)
		if actualType == "context" {
			actualType = ast.ResolveCompCallArgValue(root, arg, invokerAncestors, paramRef).Type
		}
		if actualType == "" || actualType == "param" || actualType == expectedType {
			continue
		}

		wrongType = appendUniqueStrings(wrongType, argName)
	}
	if len(wrongType) > 0 {
		return "Wrong argument type", wrongArgTypeNamesMsg(wrongType), true
	}

	if value.Bound != nil {
		for _, arg := range explicitArgs {
			name := ast.GetArgNameFromCompCallArg(arg)
			if value.Bound.Arg(name) == nil {
				continue
			}
			return "Duplicate argument", duplicateArgumentMsg(name, value.Raw), true
		}
	}

	return "", "", false
}
