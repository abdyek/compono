package html

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

func isRuleName(node ast.Node, name string) bool {
	if isRuleNil(node) {
		return false
	}
	if node.Rule().Name() != name {
		return false
	}
	return true
}

func isRuleNameOneOf(node ast.Node, names []string) bool {
	if isRuleNil(node) {
		return false
	}
	return util.InSliceString(node.Rule().Name(), names)
}

func isRuleNil(node ast.Node) bool {
	if node.Rule() == nil {
		return true
	}
	return false
}

func findChildByRuleName(children []ast.Node, name string) ast.Node {
	return findChild(children, func(child ast.Node) bool {
		if isRuleNil(child) {
			return false
		}
		if child.Rule().Name() != name {
			return false
		}
		return true
	})
}

func filterChildren(children []ast.Node, filter func(ast.Node) bool) []ast.Node {
	filtered := []ast.Node{}
	for _, child := range children {
		if filter(child) {
			filtered = append(filtered, child)
		}
	}
	return filtered
}

func findChild(children []ast.Node, filter func(ast.Node) bool) ast.Node {
	filtered := filterChildren(children, filter)
	if len(filtered) > 0 {
		return filtered[0]
	}
	return nil
}
