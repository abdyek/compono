package html

import "github.com/umono-cms/compono/ast"

type renderableNode interface {
	Condition(ast.Node) bool
	Render(ast.Node) string
}
