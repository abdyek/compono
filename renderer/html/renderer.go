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
		newParamRefInRootContent(r),
		newParamRefInLocalCompDefOfRoot(r),
		newParamRefInGlobalCompDef(r),
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
	rn := r.findRenderable(nil, node)
	if rn != nil {
		return renderNode(rn, nil, node, rn.Render)
	}
	return ""
}

func (r *renderer) renderChildren(invoker renderableNode, children []ast.Node) string {
	result := ""
	for _, child := range children {
		re := r.findRenderable(invoker, child)
		if re != nil {
			result += renderNode(re, invoker, child, re.Render)
		}
	}
	return result
}

func (r *renderer) findRenderable(invoker renderableNode, node ast.Node) renderableNode {
	for _, rn := range r.renderableNodes {
		if cond := rn.Condition(invoker, node); cond {
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
