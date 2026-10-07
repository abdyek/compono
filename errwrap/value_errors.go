package errwrap

import (
	"github.com/umono-cms/compono/ast"
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
