package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
)

// NOTE: builtinComponent can be improved. Now, It is enough.
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

func (_ *link) Render(invoker renderableNode, node ast.Node) string {
	newTabStr := ""
	newTab, ok := getBoolArgValue(node, "new-tab")
	if ok && newTab {
		newTabStr = ` target="_blank" rel="noopener noreferrer"`
	}
	return "<a href=\"" + getArgValueWithDefa(node, "url", "url") + "\"" + newTabStr + ">" + getArgValueWithDefa(node, "text", "") + "</a>"
}

func getArgValueWithDefa(compCall ast.Node, name string, defa string) string {
	value, ok := getArgValue(compCall, name)
	if !ok {
		return defa
	}
	return value
}

func getArgValue(compCall ast.Node, name string) (string, bool) {
	compCallArgs := findNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgs == nil {
		return "", false
	}
	compCallArg := findNode(compCallArgs.Children(), func(node ast.Node) bool {
		argName := findNodeByRuleName(node.Children(), "comp-call-arg-name")
		if strings.TrimSpace(string(argName.Raw())) == name {
			return true
		}
		return false
	})
	if compCallArg == nil {
		return "", false
	}
	argValue := findNodeByRuleName(findNode(findNodeByRuleName(compCallArg.Children(), "comp-call-arg-type").Children(), func(node ast.Node) bool {
		return isRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg"})
	}).Children(), "comp-call-arg-value")
	if argValue == nil {
		return "", false
	}
	return html.EscapeString(strings.TrimSpace(string(argValue.Raw()))), true
}

func getBoolArgValue(compCall ast.Node, name string) (bool, bool) {
	value, ok := getArgValue(compCall, name)
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
