package html

import (
	"github.com/umono-cms/compono/ast"
)

// invokerChain returns the node of the invoker followed by its invoker ancestors.
func invokerChain(invoker renderableNode) []ast.Node {
	if invoker == nil {
		return []ast.Node{}
	}
	return append([]ast.Node{invoker.Node()}, getAncestorsByInvoker(invoker)...)
}

// resolveBuiltinArg resolves an argument given to the built-in rendered by the
// node, explicitly or bound to the component value the node renders.
func resolveBuiltinArg(r *renderer, invoker renderableNode, node ast.Node, name string) (ast.ResolvedValue, bool) {
	return ast.ResolveFrameArg(r.root, node, name, invokerChain(invoker))
}

// resolveBuiltinParam resolves a parameter of the built-in rendered by the node,
// falling back to its default value.
func resolveBuiltinParam(r *renderer, invoker renderableNode, node ast.Node, name string) ast.ResolvedValue {
	return ast.ResolveFrameParam(r.root, node, name, invokerChain(invoker))
}

func getAncestorsByInvoker(rn renderableNode) []ast.Node {
	invoker := rn.Invoker()
	if invoker == nil {
		return []ast.Node{}
	}
	return append([]ast.Node{invoker.Node()}, getAncestorsByInvoker(invoker)...)
}
