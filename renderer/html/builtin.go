package html

import (
	"html"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

type builtinComponent interface {
	New() builtinComponent
	Name() string
	Render(invoker renderableNode, node ast.Node) string
}

type link struct {
	renderer *renderer
}

func newLink(rend *renderer) builtinComponent {
	return &link{
		renderer: rend,
	}
}

func (l *link) New() builtinComponent {
	return newLink(l.renderer)
}

func (_ *link) Name() string {
	return "LINK"
}

func (l *link) Render(invoker renderableNode, node ast.Node) string {
	newTabStr := ""
	newTab, ok := getBoolArgValue(l.renderer, invoker, node, "new-tab")
	if ok && newTab {
		newTabStr = ` target="_blank" rel="noopener noreferrer"`
	}
	ariaCurrentStr := ""
	current, ok := getArgValue(l.renderer, invoker, node, "current")
	if ok && current != "" {
		ariaCurrentStr = ` aria-current="` + current + `"`
	}
	attrs := l.renderer.callAttributeHook(l.Name(), linkConflictAttributes)
	return "<a href=\"" + util.NormalizeURL(getArgValueWithDefa(l.renderer, invoker, node, "url", "url")) + "\"" + newTabStr + ariaCurrentStr + formatHookAttributes(attrs) + ">" + getArgValueWithDefa(l.renderer, invoker, node, "text", "") + "</a>"
}

var linkConflictAttributes = map[string]bool{
	"href":         true,
	"target":       true,
	"rel":          true,
	"aria-current": true,
}

func getArgValueWithDefa(r *renderer, invoker renderableNode, compCall ast.Node, name string, defa string) string {
	value, ok := getArgValue(r, invoker, compCall, name)
	if !ok {
		return defa
	}
	return value
}

func getArgValue(r *renderer, invoker renderableNode, compCall ast.Node, name string) (string, bool) {
	resolved, ok := resolveBuiltinArg(r, invoker, compCall, name)
	if !ok || resolved.IsZero() || resolved.Type == "array" || resolved.Type == "record" || resolved.MissingContextKey != "" {
		return "", false
	}

	return html.EscapeString(util.TrimWhitespace(resolved.Raw)), true
}

func getBoolArgValue(r *renderer, invoker renderableNode, callNode ast.Node, name string) (bool, bool) {
	value, ok := getArgValue(r, invoker, callNode, name)
	if !ok {
		return false, false
	}
	if value == "false" {
		return false, true
	}
	if value == "true" {
		return true, true
	}
	return false, false
}
