package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
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
	ew.wrapInvalidParamRef(root)
	ew.wrapInvalidCompCall(root)
}

func (ew *errorWrapper) wrapInfiniteCompCall(root ast.Node) {
	ew.detectInfiniteCompCall(root, []string{})
}

func (ew *errorWrapper) detectInfiniteCompCall(node ast.Node, callStack []string) {

	ruleName := node.Rule().Name()

	if ruleName == "block-comp-call" || ruleName == "inline-comp-call" {

		block := false
		if ruleName == "block-comp-call" {
			block = true
		}

		compCallName := ew.getCompCallName(node)
		if compCallName == "" {
			return
		}

		// TODO: This is an ugly hack for built-in components. Improve it later.
		if util.InSliceString(compCallName, []string{"LINK"}) {
			return
		}

		if ew.isInCallStack(compCallName, callStack) {
			ew.wrapWithErr(node, "Infinite component call", "The call to component **"+compCallName+"** creates an infinite loop and was skipped.", block)
			return
		}

		compDef := ew.findCompDef(node, compCallName)
		if compDef == nil {
			ew.wrapWithErr(node, "Unknown component", "The component **"+compCallName+"** is not defined or not registered.", block)
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

func (ew *errorWrapper) wrapInvalidParamRef(root ast.Node) {
	paramRefs := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleName(node, "param-ref")
	})
	for _, pr := range paramRefs {
		compDefContent := ast.FindNode(ast.GetAncestors(pr), func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def-content", "global-comp-def-content"})
		})

		if compDefContent == nil {
			ew.wrapWithErr(pr, "Invalid parameter usage", "Parameters cannot be used in the root context.", false)
			continue
		}

		if ast.IsRuleName(compDefContent, "local-comp-def-content") {
			parRefNa := strings.TrimSpace(string(ast.FindNodeByRuleName(pr.Children(), "param-ref-name").Raw()))

			localCompDef := ast.FindNodeByRuleName(ast.GetAncestors(pr), "local-comp-def")
			localCompDefHead := ast.FindNodeByRuleName(localCompDef.Children(), "local-comp-def-head")
			compParams := ast.FindNodeByRuleName(localCompDefHead.Children(), "comp-params")

			title := "Unknown parameter"
			msg := "The parameter **" + parRefNa + "** is not defined for this component."

			if compParams == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}

			compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
				compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
				if strings.TrimSpace(string(compParamName.Raw())) == parRefNa {
					return true
				}
				return false
			})

			if compParam == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}
		}

		if ast.IsRuleName(compDefContent, "global-comp-def-content") {
			parRefNa := strings.TrimSpace(string(ast.FindNodeByRuleName(pr.Children(), "param-ref-name").Raw()))

			globalCompDef := ast.FindNodeByRuleName(ast.GetAncestors(pr), "global-comp-def")
			globalCompDefHead := ast.FindNodeByRuleName(globalCompDef.Children(), "global-comp-def-head")

			title := "Unknown parameter"
			msg := "The parameter **" + parRefNa + "** is not defined for this component."

			if globalCompDefHead == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}

			compParams := ast.FindNodeByRuleName(globalCompDefHead.Children(), "comp-params")
			if compParams == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}

			compParam := ast.FindNode(compParams.Children(), func(cp ast.Node) bool {
				compParamName := ast.FindNodeByRuleName(cp.Children(), "comp-param-name")
				if strings.TrimSpace(string(compParamName.Raw())) == parRefNa {
					return true
				}
				return false
			})

			if compParam == nil {
				ew.wrapWithErr(pr, title, msg, false)
				continue
			}
		}
	}
}

func (ew *errorWrapper) wrapInvalidCompCall(root ast.Node) {
	inlineCompCalls := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleName(node, "inline-comp-call")
	})

	for _, icc := range inlineCompCalls {
		compCallName := ew.getCompCallName(icc)
		compDef := ew.findCompDef(icc, compCallName)
		compDefContent := ew.getCompDefContent(compDef)
		childrenCount := len(compDefContent.Children())
		if childrenCount == 0 {
			continue
		}
		p := ast.FindNodeByRuleName(compDefContent.Children(), "p")
		if childrenCount > 1 || p == nil {
			ew.wrapWithErr(icc, "Invalid component usage", "The component **"+compCallName+"** is a block component and cannot be used inline.", false)
		}
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
