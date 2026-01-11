package html

import (
	"github.com/umono-cms/compono/ast"
)

func getAncestorsByInvoker(rn renderableNode) []ast.Node {
	invoker := rn.Invoker()
	if invoker == nil {
		return []ast.Node{}
	}
	return append([]ast.Node{invoker.Node()}, getAncestorsByInvoker(invoker)...)
}
