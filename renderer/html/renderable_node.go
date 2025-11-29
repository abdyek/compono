package html

import (
	"github.com/umono-cms/compono/ast"
)

type renderableNode interface {
	Condition(ast.Node) bool
	Render(ast.Node) string

	Invoker() renderableNode
	SetInvoker(renderableNode)

	Node() ast.Node
	SetNode(ast.Node)
}

type baseRenderable struct {
	invoker renderableNode
	node    ast.Node
}

func (br *baseRenderable) Invoker() renderableNode {
	return br.invoker
}

func (br *baseRenderable) SetInvoker(invoker renderableNode) {
	br.invoker = invoker
}

func (br *baseRenderable) Node() ast.Node {
	return br.node
}

func (br *baseRenderable) SetNode(node ast.Node) {
	br.node = node
}

func renderNode(rn renderableNode, invoker renderableNode, node ast.Node, render func(ast.Node) string) string {
	rn.SetInvoker(invoker)
	rn.SetNode(node)
	return render(node)
}
