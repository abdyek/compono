package html

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

func isRuleName(node ast.Node, name string) bool {
	rule := node.Rule()
	if rule != nil && rule.Name() == name {
		return true
	}
	return false
}

func inRuleName(node ast.Node, names []string) bool {
	rule := node.Rule()
	if rule == nil {
		return false
	}
	return util.InSliceString(rule.Name(), names)
}
