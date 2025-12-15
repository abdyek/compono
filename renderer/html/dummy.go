package html

import (
	"github.com/umono-cms/compono/ast"
)

// dummy is a minimal renderableNode used only to reproduce
// the recursive invoker traversal causing stack overflow.
type dummy struct {
	baseRenderable
}

func newDummy() renderableNode {
	return &dummy{}
}

func (_ *dummy) New() renderableNode {
	return newDummy()
}

// For local components of global components
func (_ *dummy) Condition(invoker renderableNode, node ast.Node) bool {
	if !isRuleName(node, "param-ref") {
		return false
	}
	localCompDef := findNodeByRuleName(getAncestors(node), "local-comp-def")
	if localCompDef == nil {
		return false
	}
	globalCompDef := findNodeByRuleName(getAncestors(node), "global-comp-def")
	if globalCompDef == nil {
		return false
	}
	return true
}

func (d *dummy) Render() string {
	_ = getAncestorsByInvoker(d)
	return ""
}
