package html

import (
	"html"
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
)

var codeBlockLangPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9+#._-]*$`)

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
	return ast.IsRuleName(node, "code-block")
}

func (cb *codeBlock) Render() string {
	lang := ""
	cbl := ast.FindNodeByRuleName(cb.Node().Children(), "code-block-lang")
	if cbl != nil {
		lang = strings.TrimSpace(string(cbl.Raw()))
	}
	content := cb.renderer.renderChildren(cb, cb.Node().Children())
	if lang != "" && (len(lang) > 32 || !codeBlockLangPattern.MatchString(lang)) {
		return cb.renderer.recordDiagnostic(cb.Node(), "Invalid code block language", "The language **"+lang+"** is not valid. A language name contains only lowercase ASCII letters, digits and **+ # . _ -**, starts with a letter or digit and is at most 32 bytes.", true)
	}
	if strings.Trim(content, " \t\n") == "" {
		return ""
	}
	if lang != "" {
		return `<pre><code data-lang="` + html.EscapeString(lang) + `">` + content + `</code></pre>`
	}
	return `<pre><code>` + content + `</code></pre>`
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
	return ast.IsRuleName(node, "code-block-content")
}

func (cbc *codeBlockContent) Render() string {
	return cbc.renderer.renderChildren(cbc, cbc.Node().Children())
}

type inlineCode struct {
	baseRenderable
	renderer *renderer
}

func newInlineCode(rend *renderer) renderableNode {
	return &inlineCode{
		renderer: rend,
	}
}

func (ic *inlineCode) New() renderableNode {
	return newInlineCode(ic.renderer)
}

func (_ *inlineCode) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "inline-code")
}

func (ic *inlineCode) Render() string {
	content := ic.renderer.renderChildren(ic, ic.Node().Children())
	return `<code style="white-space: pre">` + content + "</code>"
}

type inlineCodeContent struct {
	baseRenderable
	renderer *renderer
}

func newInlineCodeContent(rend *renderer) renderableNode {
	return &inlineCodeContent{
		renderer: rend,
	}
}

func (icc *inlineCodeContent) New() renderableNode {
	return newInlineCodeContent(icc.renderer)
}

func (_ *inlineCodeContent) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "inline-code-content")
}

func (icc *inlineCodeContent) Render() string {
	return icc.renderer.renderChildren(icc, icc.Node().Children())
}
