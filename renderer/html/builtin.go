package html

import (
	"html"
	"strings"

	"github.com/umono-cms/compono/ast"
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
	compCallArgs := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgs == nil {
		return "", false
	}
	compCallArg := ast.FindNode(compCallArgs.Children(), func(node ast.Node) bool {
		argName := ast.FindNodeByRuleName(node.Children(), "comp-call-arg-name")
		return strings.TrimSpace(string(argName.Raw())) == name
	})
	if compCallArg == nil {
		return "", false
	}
	compCallArgType := ast.FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	if compCallArgType == nil {
		return "", false
	}
	typedArg := ast.FindNode(compCallArgType.Children(), func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg"})
	})
	if typedArg == nil {
		return "", false
	}
	argValue := ast.FindNodeByRuleName(typedArg.Children(), "comp-call-arg-value")
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
