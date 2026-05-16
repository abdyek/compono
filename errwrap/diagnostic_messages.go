package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

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

func invalidBuiltinCompCallSchemaMsg(ctx *wrapContext, node ast.Node) string {
	mismatchedArgNames := getBuiltinSchemaMismatchArgNames(ctx, node)
	compName := getBuiltinSchemaMismatchTargetName(ctx, node)
	if compName == "" {
		compName = getCompCallNameStr(node)
	}
	if len(mismatchedArgNames) == 0 {
		return "One or more arguments do not match the schema of the built-in component **" + compName + "**."
	}

	if len(mismatchedArgNames) == 1 {
		return "The parameter **" + mismatchedArgNames[0] + "** does not match the schema of the built-in component **" + compName + "**."
	}

	return "The parameters **" + strings.Join(mismatchedArgNames, "**, **") + "** do not match the schema of the built-in component **" + compName + "**."
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

func unknownContextKeyMsg(key string) string {
	return "The key **" + key + "** is not injected."
}

func unknownRecordKeyMsg(key string) string {
	return "The key **" + key + "** is not defined in this record."
}

func paramArrayIndexOutOfRangeMsg(paramName string) string {
	return "The index used for parameter **" + paramName + "** is out of range."
}

func contextArrayIndexOutOfRangeMsg() string {
	return "The index used for this context value is out of range."
}

func invalidParamIndexAccessMsg(paramName string) string {
	return "The parameter **" + paramName + "** is not an array and cannot be indexed."
}

func invalidParamKeyAccessMsg(paramName string) string {
	return "The parameter **" + paramName + "** is not a record and cannot be accessed by key."
}

func invalidContextIndexAccessMsg(key string) string {
	return "The context value **" + key + "** is not an array and cannot be indexed."
}

func invalidContextKeyAccessMsg(key string) string {
	return "The context value **" + key + "** is not a record and cannot be accessed by key."
}

func directParamArrayUsageMsg(paramName string) string {
	return "The parameter **" + paramName + "** is an array and cannot be rendered directly."
}

func directParamRecordUsageMsg(paramName string) string {
	return "The parameter **" + paramName + "** is a record and cannot be rendered directly."
}
