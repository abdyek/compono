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
		wrapRules: wrapRules(),
	}
}

type errorWrapper struct {
	wrapRules []wrapRule
}

func (ew *errorWrapper) Wrap(root ast.Node) {
	ctx := &wrapContext{
		root:           root,
		compCallChains: ew.getCompCallChains(root),
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
Outer:
	for _, wr := range ew.wrapRules {
		for _, cond := range wr.conditions {
			if !cond(ctx, node) {
				continue Outer
			}
		}
		ew.wrapWithErr(node, wr.title(ctx, node), wr.message(ctx, node), wr.block(ctx, node))
		return true
	}

	return false
}

func (ew *errorWrapper) getCompCallChains(root ast.Node) [][]ast.Node {
	rootContent := ast.FindNodeByRuleName(root.Children(), "root-content")
	compCalls := ast.FilterNodesInTree(rootContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	chains := [][]ast.Node{}

	for _, compCall := range compCalls {
		chain := []ast.Node{}
		addLinkToChain(root, &chain, compCall)
		chains = append(chains, chain)
	}

	return chains
}

func addLinkToChain(root ast.Node, chain *[]ast.Node, compCall ast.Node) {
	stop := false
	for _, existing := range *chain {
		if existing == compCall {
			stop = true
		}
	}

	*chain = append(*chain, compCall)
	if stop {
		return
	}

	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return
	}

	compDef := findCompDef(root, compCall, compCallName)
	if compDef == nil {
		return
	}

	compCalls := ast.FilterNodesInTree(compDef, func(child ast.Node) bool {
		return ast.IsRuleNameOneOf(child, []string{"block-comp-call", "inline-comp-call"})
	})

	for _, cc := range compCalls {
		addLinkToChain(root, chain, cc)
	}
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

	return errNode
}
