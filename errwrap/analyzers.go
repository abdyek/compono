package errwrap

import "github.com/umono-cms/compono/ast"

type wrapContext struct {
	root               ast.Node
	compCallChains     [][]ast.Node
	compCallCycleCache map[ast.Node]bool
	paramCycleClosers  map[ast.Node]string
	callReplacements   map[ast.Node]ast.Node
}

type conditionAnalyzer struct {
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
