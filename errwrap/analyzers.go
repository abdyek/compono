package errwrap

import "github.com/umono-cms/compono/ast"

type wrapContext struct {
	root               ast.Node
	compCallChains     [][]ast.Node
	compCallCycleCache map[ast.Node]bool
	paramCycleClosers  map[ast.Node]string
	bindingIssues      map[ast.Node]diagnostic
}

type conditionAnalyzer struct {
	conditions []func(ctx *wrapContext, node ast.Node) bool
	title      func(ctx *wrapContext, node ast.Node) string
	message    func(ctx *wrapContext, node ast.Node) string
	block      func(ctx *wrapContext, node ast.Node) bool
}

func diagnosticAnalyzers() []diagnosticAnalyzer {
	return []diagnosticAnalyzer{
		invalidParamDef(),
		infiniteBlockCompCallByItself(),
		infiniteInlineCompCallByItself(),
		infiniteCompCallByChain(),
		infiniteCompCallByParam(),
		infiniteParamCompCallByChain(),
		unknownCompCall(),
		unknownCompParamCall(),
		blockCompInsideInline(),
		undefinedParam(),
		missingArg(),
		wrongImageArgType(),
		invalidImage(),
		invalidBuiltinCompCallSchema(),
		wrongArgType(),
		boundArgumentAnalyzer{},
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
