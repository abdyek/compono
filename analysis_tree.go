package compono

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/parser"
	"github.com/umono-cms/compono/rule"
)

// analysisTree parses source for Analyze. A source that starts with a
// parameter line is parsed as a global component definition with an empty
// name, placed in the global scope of an empty root; any other source is
// parsed as a converted source. top is the node that holds the content and
// the local definitions, root is the root of the tree.
func analysisTree(source []byte) (top, root ast.Node) {
	p := parser.DefaultParser(logger.NewLogger())

	global := ast.DefaultEmptyNode()
	global.SetRule(rule.NewGlobalCompDef())
	global = p.Parse(source, global)
	if ast.FindNodeByRuleName(global.Children(), "global-comp-def-head") == nil {
		root = p.Parse(source, ast.DefaultRootNode())
		return root, root
	}

	name := ast.DefaultEmptyNode()
	name.SetRule(rule.NewGlobalCompName())
	name.SetParent(global)
	global.SetChildren(append([]ast.Node{name}, global.Children()...))

	root = p.Parse([]byte{}, ast.DefaultRootNode())
	wrapper := ast.DefaultEmptyNode()
	wrapper.SetRule(rule.NewGlobalCompDefWrapper())
	wrapper.SetParent(root)
	wrapper.SetChildren([]ast.Node{global})
	global.SetParent(wrapper)
	root.SetChildren(append(root.Children(), wrapper))
	return global, root
}

// attachBuiltins adds the built-in components to root, as Convert does, so
// the error wrapper can resolve calls to them.
func attachBuiltins(root ast.Node) {
	builtinWrapper := ast.DefaultEmptyNode()
	builtinWrapper.SetRule(rule.NewDynamic("builtin-comp-wrapper"))
	builtinWrapper.SetParent(root)
	builtinWrapper.SetChildren(builtin.BuildASTNodes(builtinWrapper))
	root.SetChildren(append(root.Children(), builtinWrapper))
}
