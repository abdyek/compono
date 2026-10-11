package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/errwrap"
	"github.com/umono-cms/compono/util"
)

type linkElement struct {
	baseRenderable
	renderer *renderer
}

func newLinkElement(rend *renderer) renderableNode {
	return &linkElement{
		renderer: rend,
	}
}

func (l *linkElement) New() renderableNode {
	return newLinkElement(l.renderer)
}

func (_ *linkElement) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "link")
}

func (l *linkElement) Render() string {
	linkText := ast.FindNodeByRuleName(l.Node().Children(), "link-text")
	linkURL := ast.FindNodeByRuleName(l.Node().Children(), "link-url")

	if inlineCompCall := findFirstInlineCompCall(linkText, linkURL); inlineCompCall != nil {
		compCallName := ast.FindNodeByRuleName(inlineCompCall.Children(), "comp-call-name")
		name := util.TrimWhitespace(string(compCallName.Raw()))
		return l.renderer.recordDiagnostic(l.Node(), "Invalid component usage", "The component **"+name+"** cannot be used inside a link.", false)
	}

	text := ""
	if linkText != nil {
		text = util.TrimWhitespace(l.renderer.renderChildren(l, linkText.Children()))
	}
	if isEmptyRendered(text) {
		return text + l.renderer.recordDiagnostic(l.Node(), "Empty text", "The link text cannot be empty.", false)
	}

	url := ""
	if linkURL != nil {
		url = l.renderURL(linkURL)
	}
	if diagnosticMarkerPattern.MatchString(url) {
		return strings.Join(diagnosticMarkerPattern.FindAllString(url, -1), "")
	}
	if util.TrimWhitespace(url) == "" {
		return l.renderer.recordDiagnostic(l.Node(), "Empty URL", "The link URL cannot be empty.", false)
	}

	url = util.NormalizeURL(url)

	if scheme, disallowed := util.DisallowedURLScheme(url); disallowed {
		title, message := errwrap.DisallowedURLSchemeError(scheme)
		return l.renderer.recordDiagnostic(l.Node(), title, message, false)
	}

	return `<a href="` + url + `">` + text + `</a>`
}

func (l *linkElement) renderURL(linkURL ast.Node) string {
	if !linkURL.HasChildren() {
		return html.EscapeString(strings.TrimSpace(string(linkURL.Raw())))
	}

	return strings.TrimSpace(l.renderer.renderChildren(l, linkURL.Children()))
}

func findFirstInlineCompCall(linkText, linkURL ast.Node) ast.Node {
	if linkText != nil {
		if found := ast.FilterNodesInTree(linkText, func(node ast.Node) bool {
			return ast.IsRuleName(node, "inline-comp-call")
		}); len(found) > 0 {
			return found[0]
		}
	}
	if linkURL != nil {
		if found := ast.FilterNodesInTree(linkURL, func(node ast.Node) bool {
			return ast.IsRuleName(node, "inline-comp-call")
		}); len(found) > 0 {
			return found[0]
		}
	}
	return nil
}

type linkTextElement struct {
	baseRenderable
	renderer *renderer
}

func newLinkTextElement(rend *renderer) renderableNode {
	return &linkTextElement{
		renderer: rend,
	}
}

func (lt *linkTextElement) New() renderableNode {
	return newLinkTextElement(lt.renderer)
}

func (_ *linkTextElement) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "link-text")
}

func (lt *linkTextElement) Render() string {
	return html.EscapeString(strings.TrimSpace(string(lt.Node().Raw())))
}

type linkURLElement struct {
	baseRenderable
	renderer *renderer
}

func newLinkURLElement(rend *renderer) renderableNode {
	return &linkURLElement{
		renderer: rend,
	}
}

func (lu *linkURLElement) New() renderableNode {
	return newLinkURLElement(lu.renderer)
}

func (_ *linkURLElement) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "link-url")
}

func (lu *linkURLElement) Render() string {
	return html.EscapeString(strings.TrimSpace(string(lu.Node().Raw())))
}
