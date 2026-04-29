package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

func alwaysBlock(_ *wrapContext, _ ast.Node) bool { return true }
func neverBlock(_ *wrapContext, _ ast.Node) bool  { return false }

func blockFromRuleName(_ *wrapContext, node ast.Node) bool {
	return strings.HasPrefix(node.Rule().Name(), "block-")
}

func blockForParamRef(_ *wrapContext, node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}

	pContent := ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "p-content")
	})
	if pContent == nil {
		return false
	}

	for _, child := range pContent.Children() {
		if ast.IsRuleName(child, "soft-break") {
			return true
		}
	}

	for _, child := range pContent.Children() {
		if child == node {
			continue
		}
		if ast.IsRuleName(child, "plain") && strings.TrimSpace(string(child.Raw())) == "" {
			continue
		}
		return false
	}

	return true
}

func blockUndefinedParamRef(_ *wrapContext, node ast.Node) bool {
	if !blockForParamRef(nil, node) {
		return false
	}
	refName := getParamRefNameStr(node)
	return strings.Contains(refName, "comp")
}
