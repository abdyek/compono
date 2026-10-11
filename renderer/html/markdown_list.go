package html

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

type markdownListItem struct {
	baseRenderable
	renderer *renderer
}

func newMarkdownListItem(rend *renderer) renderableNode {
	return &markdownListItem{
		renderer: rend,
	}
}

func (m *markdownListItem) New() renderableNode {
	return newMarkdownListItem(m.renderer)
}

func (_ *markdownListItem) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "list-item")
}

func (m *markdownListItem) Render() string {
	content := util.TrimWhitespace(normalizeRenderedMarkup(m.renderer.renderChildren(m, m.Node().Children())))
	if isEmptyRendered(content) {
		return content
	}
	return "<li>" + content + "</li>"
}

type markdownList struct {
	baseRenderable
	renderer *renderer
}

func newMarkdownList(rend *renderer) renderableNode {
	return &markdownList{
		renderer: rend,
	}
}

func (m *markdownList) New() renderableNode {
	return newMarkdownList(m.renderer)
}

func (_ *markdownList) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "list")
}

func (m *markdownList) Render() string {
	content := ""
	hasItem := false
	for _, child := range m.Node().Children() {
		rendered := m.renderer.renderChildren(m, []ast.Node{child})
		if strings.HasPrefix(rendered, "<li>") {
			hasItem = true
		}
		content += rendered
	}
	if hasItem {
		return "<ul>" + content + "</ul>"
	}
	return content
}
