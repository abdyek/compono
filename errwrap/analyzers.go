package errwrap

import "github.com/umono-cms/compono/ast"

type wrapContext struct {
	root ast.Node
}

type conditionAnalyzer struct {
	conditions []func(ctx *wrapContext, node ast.Node) bool
	title      func(ctx *wrapContext, node ast.Node) string
	message    func(ctx *wrapContext, node ast.Node) string
	block      func(ctx *wrapContext, node ast.Node) bool
}

func diagnosticAnalyzers() []diagnosticAnalyzer {
	return []diagnosticAnalyzer{
		invalidExpression(),
		unclosedExpression(),
		invalidParamDef(),
		unknownCompCall(),
		unknownCompParamCall(),
		blockCompInsideInline(),
		undefinedParam(),
		missingArg(),
		wrongImageArgType(),
		wrongArgType(),
		wrongBoundArgTypeInBuiltinCall(),
		paramRefInRootContent(),
		paramRefInLinkInRootContent(),
		contextRefAnalyzer{},
		undefinedParamRef(),
		undefinedParamRefInLink(),
		notCompParamCompCall(),
		undefinedParamCompCall(),
		undefinedParamArgRefInParamCompCall(),
		wrongBoundArgTypeInParamCompCall(),
	}
}
