package errwrap

import "github.com/umono-cms/compono/ast"

func isRuleName(name string) func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return ast.IsRuleName(node, name)
	}
}

func isRuleNameOneOf(names ...string) func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, names)
	}
}

func hasCompCallArgs() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return hasCompCallArgsNode(node)
	}
}

func hasCompCallArgsNode(node ast.Node) bool {
	return ast.FindNodeByRuleName(node.Children(), "comp-call-args") != nil
}

func any(conds ...func(*wrapContext, ast.Node) bool) func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		for _, cond := range conds {
			if cond(ctx, node) {
				return true
			}
		}
		return false
	}
}

func not(cond func(*wrapContext, ast.Node) bool) func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !cond(ctx, node)
	}
}
