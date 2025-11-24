package html

import (
	"io"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
)

type renderer struct {
	logger          logger.Logger
	renderableNodes []renderableNode
	root            ast.Node
	builtinCompMap  map[string]builtinComponent
}

func NewRenderer(log logger.Logger) *renderer {

	r := &renderer{
		logger: log,
	}

	r.renderableNodes = []renderableNode{
		newRoot(r),
		newRootContent(r),
		newCompCall(r),
		newNonVoidElement(r),
		newNonVoidElementContent(r),
		newPlain(r),
	}

	r.builtinCompMap = make(map[string]builtinComponent)

	builtinComps := []builtinComponent{
		NewLink(r),
	}

	for _, bc := range builtinComps {
		r.builtinCompMap[bc.Name()] = bc
	}

	return r
}

func (r *renderer) Render(writer io.Writer, root ast.Node) error {

	r.root = root

	_, err := writer.Write([]byte(r.render(root)))
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) render(node ast.Node) string {
	rn := r.findRenderable(node)
	if rn != nil {
		return rn.Render(node)
	}
	return ""
}

func (r *renderer) renderChildren(children []ast.Node) string {
	result := ""
	for _, child := range children {
		re := r.findRenderable(child)
		if re != nil {
			result += re.Render(child)
		}
	}
	return result
}

func (r *renderer) findRenderable(node ast.Node) renderableNode {
	for _, rn := range r.renderableNodes {
		if cond := rn.Condition(node); cond {
			return rn
		}
	}
	return nil
}

func (r *renderer) findLocalCompDef(srcNode ast.Node, name string) ast.Node {
	localCompDefWrapper := findNodeByRuleName(srcNode.Children(), "local-comp-def-wrapper")
	if localCompDefWrapper == nil {
		return nil
	}

	return findNode(localCompDefWrapper.Children(), func(child ast.Node) bool {
		if isRuleNil(child) {
			return false
		}

		if !isRuleName(child, "local-comp-def") {
			return false
		}

		localCompDefHead := findNodeByRuleName(child.Children(), "local-comp-def-head")
		if localCompDefHead == nil {
			return false
		}

		localCompName := findNodeByRuleName(localCompDefHead.Children(), "local-comp-name")
		if localCompName == nil {
			return false
		}

		if strings.TrimSpace(string(localCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

func (r *renderer) findGlobalCompDef(name string) ast.Node {
	globalCompDefWrapper := findNodeByRuleName(r.root.Children(), "global-comp-def-wrapper")
	if globalCompDefWrapper == nil {
		return nil
	}

	return findNode(globalCompDefWrapper.Children(), func(child ast.Node) bool {
		if isRuleNil(child) {
			return false
		}

		if !isRuleName(child, "global-comp-def") {
			return false
		}

		globalCompName := findNodeByRuleName(child.Children(), "global-comp-name")
		if globalCompName == nil {
			return false
		}

		if strings.TrimSpace(string(globalCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

func (r *renderer) findBuiltinComp(name string) builtinComponent {
	bc, ok := r.builtinCompMap[strings.TrimSpace(name)]
	if !ok {
		return nil
	}
	return bc
}
