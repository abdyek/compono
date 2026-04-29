package errwrap

import "github.com/umono-cms/compono/ast"

func infiniteBlockCompCallByItself() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(_ *wrapContext, node ast.Node) bool{
			isRuleName("block-comp-call"),
			isCalledByItself(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   alwaysBlock,
	}
}

func infiniteInlineCompCallByItself() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("inline-comp-call"),
			isCalledByItself(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   neverBlock,
	}
}

func infiniteCompCallByChain() conditionAnalyzer {
	return conditionAnalyzer{
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

func infiniteCompCallByParam() conditionAnalyzer {
	return conditionAnalyzer{
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

func infiniteParamCompCallByChain() conditionAnalyzer {
	return conditionAnalyzer{
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
