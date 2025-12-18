package html

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

func isRuleName(node ast.Node, name string) bool {
	if node.Rule().Name() != name {
		return false
	}
	return true
}

func isRuleNameOneOf(node ast.Node, names []string) bool {
	return util.InSliceString(node.Rule().Name(), names)
}

func findNodeByRuleName(nodes []ast.Node, name string) ast.Node {
	return findNode(nodes, func(node ast.Node) bool {
		return isRuleName(node, name)
	})
}

func filterNodes(nodes []ast.Node, filter func(ast.Node) bool) []ast.Node {
	filtered := []ast.Node{}
	for _, node := range nodes {
		if filter(node) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func findNode(nodes []ast.Node, filter func(ast.Node) bool) ast.Node {
	filtered := filterNodes(nodes, filter)
	if len(filtered) > 0 {
		return filtered[0]
	}
	return nil
}

func getAncestors(node ast.Node) []ast.Node {
	parent := node.Parent()
	if parent == nil {
		return []ast.Node{}
	}
	return append([]ast.Node{parent}, getAncestors(parent)...)
}

func getAncestorsByInvoker(rn renderableNode) []ast.Node {
	invoker := rn.Invoker()
	if invoker == nil {
		return []ast.Node{}
	}
	return append([]ast.Node{invoker.Node()}, getAncestorsByInvoker(invoker)...)
}

func inlineError(title, description string) string {
	return `<compono-error-inline><span slot="title">` +
		title +
		`</span><span slot="description">` +
		description +
		`</span></compono-error-inline>`
}

func blockError(title, description string) string {
	return `<compono-error-block><div slot="title">` +
		title +
		`</div><div slot="description">` +
		description +
		`</div></compono-error-block>`
}
