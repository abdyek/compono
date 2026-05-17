package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

type navigation struct {
	renderer *renderer
}

func newNavigation(rend *renderer) builtinComponent {
	return &navigation{
		renderer: rend,
	}
}

func (nav *navigation) New() builtinComponent {
	return newNavigation(nav.renderer)
}

func (_ *navigation) Name() string {
	return "NAVIGATION"
}

func (nav *navigation) Render(invoker renderableNode, node ast.Node) string {
	items := nav.resolveArg(invoker, node, "items")
	return `<compono-navigation><nav>` + nav.renderItems(items) + `</nav></compono-navigation>`
}

func (nav *navigation) resolveArg(invoker renderableNode, compCall ast.Node, name string) ast.ResolvedValue {
	arg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), name)
	if arg != nil {
		return ast.ResolveCompCallArgValue(nav.renderer.root, arg, getAncestorsByInvoker(invoker), compCall)
	}
	return ast.ResolveParamDefaultFromCompCall(nav.renderer.root, compCall, name)
}

func (nav *navigation) renderItems(value ast.ResolvedValue) string {
	items := make([]string, 0, len(value.Items))
	for _, item := range value.Items {
		if item.Type != "record" {
			continue
		}
		items = append(items, nav.renderItem(item))
	}
	return `<ul>` + strings.Join(items, "") + `</ul>`
}

func (nav *navigation) renderItem(item ast.ResolvedValue) string {
	label := nav.recordField(item, "label")
	target := nav.recordField(item, "target")

	rendered := `<li><a href="` + html.EscapeString(target) + `">` + html.EscapeString(label) + `</a>`
	if children, ok := item.Fields["children"]; ok && children.Type == "array" {
		rendered += nav.renderItems(children)
	}
	return rendered + `</li>`
}

func (_ *navigation) recordField(record ast.ResolvedValue, key string) string {
	field, ok := record.Fields[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(field.Raw)
}
