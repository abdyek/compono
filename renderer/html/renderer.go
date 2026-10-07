package html

import (
	"io"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/internal/attrhook"
	"github.com/umono-cms/compono/logger"
)

// Options holds the settings of a single Render call.
type Options struct {
	AttributeHook   attrhook.AttributeHookFunc
	ErrorStylesheet string
}

type renderer struct {
	logger          logger.Logger
	renderableNodes []renderableNode
	root            ast.Node
	builtinCompMap  map[string]builtinComponent
	errorStylesheet string
	attributeHook   attributeHookFunc
	frameStack      []frame
	attrErr         error
}

func NewRenderer(log logger.Logger) *renderer {
	r := &renderer{
		logger: log,
	}

	r.renderableNodes = []renderableNode{
		newErr(r),
		newRoot(r),
		newRootContent(r),
		newCompCall(r),
		newNonVoidElement(r),
		newNonVoidElementContent(r),
		newContextRef(r),
		newParamRefInLocalCompDef(r),
		newParamRefInGlobalCompDef(r),
		newPlain(r),
		newCodeBlock(r),
		newCodeBlockContent(r),
		newInlineCode(r),
		newInlineCodeContent(r),
		newRaw(r),
		newLinkElement(r),
		newLinkTextElement(r),
		newLinkURLElement(r),
		newBr(r),
	}

	r.builtinCompMap = make(map[string]builtinComponent)

	builtinComps := []builtinComponent{
		newLink(r),
		newImage(r),
	}

	for _, bc := range builtinComps {
		r.builtinCompMap[bc.Name()] = bc
	}

	return r
}

func (r *renderer) Render(writer io.Writer, root ast.Node, opts Options) error {
	cr := NewRenderer(r.logger)
	cr.root = root
	cr.attributeHook = opts.AttributeHook
	cr.errorStylesheet = opts.ErrorStylesheet

	out := cr.render(root)
	if cr.attrErr != nil {
		return cr.attrErr
	}

	_, err := writer.Write([]byte(out))
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) render(node ast.Node) string {
	rn := r.findRenderable(nil, node)
	if rn != nil {
		return renderNode(rn, nil, node)
	}
	return ""
}

func (r *renderer) renderChildren(invoker renderableNode, children []ast.Node) string {
	result := ""
	for _, child := range children {
		re := r.findRenderable(invoker, child)
		if re != nil {
			result += renderNode(re, invoker, child)
		}
	}
	return result
}

func (r *renderer) findRenderable(invoker renderableNode, node ast.Node) renderableNode {
	for _, rn := range r.renderableNodes {
		if cond := rn.Condition(invoker, node); cond {
			return rn.New()
		}
	}
	return nil
}

func (r *renderer) findLocalCompDef(srcNode ast.Node, name string) ast.Node {
	return ast.FindLocalCompDef(srcNode, name)
}

func (r *renderer) findGlobalCompDef(from ast.Node, name string) ast.Node {
	return ast.FindGlobalCompDef(r.root, from, name)
}

func (r *renderer) findBuiltinComp(name string) builtinComponent {
	if r.findBuiltinCompDef(name) == nil {
		return nil
	}

	bc, ok := r.builtinCompMap[strings.TrimSpace(name)]
	if !ok {
		return nil
	}
	return bc.New()
}

func (r *renderer) findBuiltinCompDef(name string) ast.Node {
	return ast.FindBuiltinCompDef(r.root, name)
}
