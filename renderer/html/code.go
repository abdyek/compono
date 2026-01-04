package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type codeBlock struct {
	baseRenderable
	renderer *renderer
}

func newCodeBlock(rend *renderer) renderableNode {
	return &codeBlock{
		renderer: rend,
	}
}

func (cb *codeBlock) New() renderableNode {
	return newCodeBlock(cb.renderer)
}

func (_ *codeBlock) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleName(node, "code-block")
}

func (cb *codeBlock) Render() string {
	langClass := "language-plaintext"
	cbl := findNodeByRuleName(cb.Node().Children(), "code-block-lang")
	if cbl != nil {
		lang := html.EscapeString(strings.TrimSpace(string(cbl.Raw())))
		if lang != "" {
			langClass = "language-" + lang
		}
	}
	return `<pre><code class="` + langClass + `">` + cb.renderer.renderChildren(cb, cb.Node().Children()) + `</code></pre>`
}

type codeBlockContent struct {
	baseRenderable
	renderer *renderer
}

func newCodeBlockContent(rend *renderer) renderableNode {
	return &codeBlockContent{
		renderer: rend,
	}
}

func (cbc *codeBlockContent) New() renderableNode {
	return newCodeBlockContent(cbc.renderer)
}

func (_ *codeBlockContent) Condition(invoker renderableNode, node ast.Node) bool {
	return isRuleName(node, "code-block-content")
}

func (cbc *codeBlockContent) Render() string {
	return cbc.renderer.renderChildren(cbc, cbc.Node().Children())
}
