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

func paramRefInLinkInRootContent() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("link"),
			isInsideRootContent(),
			hasParamRefInLink(),
		},
		title:   staticTitle("Invalid parameter usage"),
		message: paramRefInRootMsg,
		block:   neverBlock,
	}
}

func undefinedParamRefInLink() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("link"),
			not(isInsideRootContent()),
			hasUndefinedParamRefInLink(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamRefsInLinkMsg,
		block:   neverBlock,
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

func undefinedParamArgRefInParamCompCall() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			hasCompCallArgs(),
			not(isInsideRootContent()),
			func(ctx *wrapContext, node ast.Node) bool {
				return len(getUndefinedParamCompCallArgNames(ctx, node)) > 0
			},
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamArgRefsMsg,
		block:   blockForParamRef,
	}
}

func wrongBoundArgTypeInParamCompCall() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			hasCompCallArgs(),
			not(isInsideRootContent()),
			func(ctx *wrapContext, node ast.Node) bool {
				return len(getWrongTypeBoundArgNames(ctx.root, node)) > 0
			},
		},
		title: staticTitle("Wrong argument type"),
		message: func(ctx *wrapContext, node ast.Node) string {
			return wrongArgTypeNamesMsg(getWrongTypeBoundArgNames(ctx.root, node))
		},
		block: blockForParamRef,
	}
}

func getUndefinedParamCompCallArgNames(ctx *wrapContext, node ast.Node) []string {
	return appendUniqueStrings(getUndefinedParamArgRefNames(node), getUndefinedBoundArgNames(ctx.root, node)...)
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

		compDef := findEnclosingCompDef(paramRef)
		if compDef == nil {
			return false
		}

		return !util.InSliceString(refName, getCompDefParamNames(compDef))
	}
}

// getUndefinedParamArgRefNames returns the parameters referenced by the argument
// values of a call that are not defined by the component the call is written in.
func getUndefinedParamArgRefNames(call ast.Node) []string {
	compDef := findEnclosingCompDef(call)
	if compDef == nil {
		return nil
	}

	definedParams := getCompDefParamNames(compDef)
	names := []string{}
	for _, arg := range ast.GetCompCallArgsFromCompCall(call) {
		for _, paramArg := range ast.FilterNodesInTree(arg, func(node ast.Node) bool {
			return ast.IsRuleName(node, "comp-call-param-arg")
		}) {
			refName, _ := ast.GetValuePathFromRaw(string(paramArg.Raw()))
			if refName == "" || util.InSliceString(refName, definedParams) {
				continue
			}
			names = appendUniqueStrings(names, refName)
		}
	}

	return names
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

func hasParamRefInLink() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, link ast.Node) bool {
		return len(getParamRefsInLink(link)) > 0
	}
}

func hasUndefinedParamRefInLink() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, link ast.Node) bool {
		return len(getUndefinedParamRefNamesInLink(ctx, link)) > 0
	}
}

func getParamRefsInLink(link ast.Node) []ast.Node {
	return ast.FilterNodesInTree(link, func(node ast.Node) bool {
		return ast.IsRuleName(node, "param-ref")
	})
}

func getUndefinedParamRefNamesInLink(ctx *wrapContext, link ast.Node) []string {
	names := []string{}
	isUndefined := isUndefinedParamRef()

	for _, paramRef := range getParamRefsInLink(link) {
		refName := getParamRefNameStr(paramRef)
		if refName == "" || util.InSliceString(refName, names) {
			continue
		}
		if !isUndefined(ctx, paramRef) {
			continue
		}
		names = append(names, refName)
	}

	return names
}
