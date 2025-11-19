package html

import (
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/util"
)

type renderer struct {
	logger          logger.Logger
	renderableNodes []renderableNode
	root            ast.Node
}

func NewRenderer(log logger.Logger) *renderer {

	r := &renderer{
		logger: log,
	}

	r.renderableNodes = []renderableNode{
		newRoot(r),
		newRootContent(r),
		newBlockCompCall(r),
		newNonVoidElement(r),
		newNonVoidElementContent(r),
		newPlain(r),
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

func (r *renderer) isRuleName(node ast.Node, name string) bool {
	rule := node.Rule()
	if rule != nil && rule.Name() == name {
		return true
	}
	return false
}

func (r *renderer) inRuleName(node ast.Node, names []string) bool {
	rule := node.Rule()
	if rule == nil {
		return false
	}
	return util.InSliceString(rule.Name(), names)
}
