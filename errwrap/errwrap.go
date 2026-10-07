package errwrap

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
)

type ErrorWrapper interface {
	Wrap(ast.Node)
}

func DefaultErrorWrapper() ErrorWrapper {
	return &errorWrapper{
		analyzers: diagnosticAnalyzers(),
	}
}

type errorWrapper struct {
	analyzers []diagnosticAnalyzer
}

func (ew *errorWrapper) Wrap(root ast.Node) {
	ctx := &wrapContext{
		root: root,
	}

	ew.scanAndWrap(ctx, root)
}

func (ew *errorWrapper) scanAndWrap(ctx *wrapContext, node ast.Node) {
	if ew.wrap(ctx, node) {
		return
	}

	for _, child := range node.Children() {
		ew.scanAndWrap(ctx, child)
	}
}

func (ew *errorWrapper) wrap(ctx *wrapContext, node ast.Node) (wrapped bool) {
	for _, analyzer := range ew.analyzers {
		diag, ok := analyzer.Diagnose(ctx, node)
		if !ok {
			continue
		}
		ew.wrapWithErr(node, diag.title, diag.message, diag.block)
		return true
	}

	return false
}

func (ew *errorWrapper) wrapWithErr(self ast.Node, title, msg string, block bool) {
	var errNode ast.Node
	if block {
		errNode = ew.createBlockError(self, title, msg)
	} else {
		errNode = ew.createInlineError(self, title, msg)
	}

	self.SetRule(errNode.Rule())
	self.SetChildren(errNode.Children())
	self.SetRaw(errNode.Raw())
}

func (ew *errorWrapper) createBlockError(node ast.Node, title, msg string) ast.Node {
	return ew.createError("block-error", node, title, msg)
}

func (ew *errorWrapper) createInlineError(node ast.Node, title, msg string) ast.Node {
	return ew.createError("inline-error", node, title, msg)
}

func (ew *errorWrapper) createError(errRuleName string, node ast.Node, title, msg string) ast.Node {
	err := rule.NewDynamic(errRuleName)
	errTitle := rule.NewDynamic("error-title")
	errMsg := rule.NewDynamic("error-message")
	self := rule.NewDynamic("self")

	errNode := ast.DefaultEmptyNode()
	errNode.SetRule(err)

	errTitleNode := ast.DefaultEmptyNode()
	errTitleNode.SetRule(errTitle)
	errTitleNode.SetParent(errNode)
	errTitleNode.SetRaw([]byte(title))

	errMsgNode := ast.DefaultEmptyNode()
	errMsgNode.SetRule(errMsg)
	errMsgNode.SetParent(errNode)
	errMsgNode.SetRaw([]byte(msg))

	selfNode := ast.DefaultEmptyNode()
	selfNode.SetRule(self)
	selfNode.SetParent(errNode)
	selfNode.SetChildren(node.Children())

	errNode.SetChildren([]ast.Node{
		errTitleNode,
		errMsgNode,
		selfNode,
	})
	errNode.SetRange(node.Range())

	return errNode
}

func getMissingContextKeyInBoundArgs(root ast.Node, arg ast.Node) string {
	for _, boundArgs := range ast.FilterNodesInTree(arg, func(node ast.Node) bool {
		return ast.IsRuleName(node, "comp-bound-args")
	}) {
		for _, contextArg := range ast.FilterNodesInTree(boundArgs, func(node ast.Node) bool {
			return ast.IsRuleName(node, "comp-call-context-arg")
		}) {
			if key := ast.ResolveContextReferenceValue(root, contextArg).MissingContextKey; key != "" {
				return key
			}
		}
	}
	return ""
}
