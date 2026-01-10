package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
)

type ErrorWrapper interface {
	Wrap(ast.Node)
}

func DefaultErrorWrapper() ErrorWrapper {
	return &errorWrapper{}
}

type errorWrapper struct {
	root ast.Node
}

func (ew *errorWrapper) Wrap(root ast.Node) {
	ew.root = root
	ew.wrapInfiniteCompCall(root)
}

func (ew *errorWrapper) wrapInfiniteCompCall(root ast.Node) {
	ew.detectInfiniteCompCall(root, []string{})
}

func (ew *errorWrapper) detectInfiniteCompCall(node ast.Node, callStack []string) {

	ruleName := node.Rule().Name()

	// TODO: separate as block and inline
	if ruleName == "block-comp-call" || ruleName == "inline-comp-call" {
		compCallName := ew.getCompCallName(node)
		if compCallName == "" {
			return
		}

		if ew.isInCallStack(compCallName, callStack) {
			ew.markAsInfiniteLoop(node, callStack, compCallName)
			return
		}

		compDef := ew.findCompDef(node, compCallName)
		if compDef == nil {
			// TODO: Move the unknown component error here from the renderer
			return
		}

		newStack := append(callStack, compCallName)
		compDefContent := ew.getCompDefContent(compDef)
		if compDefContent != nil {
			ew.detectInfiniteCompCall(compDefContent, newStack)
		}
	}

	for _, child := range node.Children() {
		ew.detectInfiniteCompCall(child, callStack)
	}
}

func (ew *errorWrapper) getCompCallName(node ast.Node) string {
	compCallNameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
	if compCallNameNode != nil {
		return strings.TrimSpace(string(compCallNameNode.Raw()))
	}
	return ""
}

func (ew *errorWrapper) isInCallStack(name string, callStack []string) bool {
	for _, n := range callStack {
		if n == name {
			return true
		}
	}
	return false
}

func (ew *errorWrapper) findCompDef(compCallNode ast.Node, name string) ast.Node {

	globalCompDefAnc := ast.FindNode(ast.GetAncestors(compCallNode), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := ew.root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := ast.FindLocalCompDef(localCompDefSrc, name)
	if localCompDef != nil {
		return localCompDef
	}

	globalCompDef := ast.FindGlobalCompDef(ew.root, name)
	if globalCompDef != nil {
		return globalCompDef
	}

	return nil
}

func (ew *errorWrapper) getCompDefContent(compDef ast.Node) ast.Node {
	for _, child := range compDef.Children() {
		if child.Rule() == nil {
			continue
		}
		ruleName := child.Rule().Name()
		if ruleName == "local-comp-def-content" || ruleName == "global-comp-def-content" {
			return child
		}
	}
	return nil
}

func (ew *errorWrapper) markAsInfiniteLoop(node ast.Node, callStack []string, compName string) {

	// TODO: Refactor
	err := rule.NewDynamic("error")
	errTitle := rule.NewDynamic("error-title")
	errMsg := rule.NewDynamic("error-message")
	self := rule.NewDynamic("self")

	errNode := ast.DefaultEmptyNode()
	errNode.SetRule(err)

	errTitleNode := ast.DefaultEmptyNode()
	errTitleNode.SetRule(errTitle)
	errTitleNode.SetParent(errNode)
	errTitleNode.SetRaw([]byte("infinite loop"))

	errMsgNode := ast.DefaultEmptyNode()
	errMsgNode.SetRule(errMsg)
	errMsgNode.SetParent(errNode)
	// TODO: Complete it
	errMsgNode.SetRaw([]byte("intinite loop more details"))

	selfNode := ast.DefaultEmptyNode()
	selfNode.SetRule(self)
	selfNode.SetParent(errNode)
	selfNode.SetChildren(node.Children())

	errNode.SetChildren([]ast.Node{
		errTitleNode,
		errMsgNode,
		selfNode,
	})

	node.SetRule(err)
	node.SetChildren(errNode.Children())
	node.SetRaw(errNode.Raw())
}
