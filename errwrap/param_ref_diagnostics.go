package errwrap

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

func paramRefInRootContent() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			isInsideRootContent(),
		},
		title:   staticTitle("Invalid parameter usage"),
		message: paramRefInRootMsg,
		block:   neverBlock,
	}
}

func undefinedParamRef() conditionAnalyzer {
	return conditionAnalyzer{
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

func undefinedParamCompCall() conditionAnalyzer {
	return conditionAnalyzer{
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

func notCompParamCompCall() conditionAnalyzer {
	return conditionAnalyzer{
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
