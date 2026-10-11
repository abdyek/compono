package html

import (
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/errwrap"
	"github.com/umono-cms/compono/util"
)

// list is the LIST built-in. It renders an <ol> or <ul> of the records of its
// "items" parameter, wrapping each rendered item component in <li>.
type list struct {
	renderer *renderer
}

func newList(rend *renderer) builtinComponent {
	return &list{
		renderer: rend,
	}
}

func (l *list) New() builtinComponent {
	return newList(l.renderer)
}

// Name returns the name of the built-in component.
func (_ *list) Name() string {
	return "LIST"
}

// Render renders the LIST built-in as an <ol> when ordered is true, otherwise
// as a <ul>. Each non-empty item rendered as the "item" component is wrapped
// in <li>; items that render as empty are emitted as is.
func (l *list) Render(invoker renderableNode, node ast.Node) string {
	r := l.renderer

	items := resolveBuiltinParam(r, invoker, node, "items")
	itemValue := resolveBuiltinParam(r, invoker, node, "item")
	if itemValue.Type != "comp" || itemValue.Raw == "" {
		return ""
	}

	compDef := ast.FindCompDefInScope(r.root, itemValue.Scope, node, itemValue.Raw)
	if compDef == nil {
		return r.recordDiagnostic(node, "Unknown component", "The component **"+itemValue.Raw+"** is not defined or not registered.", true)
	}

	inline := !errwrap.IsBlockComponent(compDef)

	tag := "ul"
	if ordered, ok := getBoolArgValue(r, invoker, node, "ordered"); ok && ordered {
		tag = "ol"
	}
	attrs := r.callAttributeHook(l.Name(), map[string]bool{})

	r.pushBuiltinFrame(l.Name(), node)
	defer r.popFrame()

	out := ""
	wroteItem := false
	for i, element := range items.Items {
		if element.Type != "record" {
			out += r.recordDiagnostic(node, "Invalid list item", "The item at index **"+strconv.Itoa(i)+"** is not a record.", true)
			continue
		}
		frame := ast.NewCompValueFrameWithArgs(itemValue, node, map[string]ast.ResolvedValue{"item": element})
		content := renderCompValue(r, newPassthroughRenderable(frame, invoker), itemValue, inline)
		if isEmptyRendered(content) {
			out += content
			continue
		}
		out += "<li>" + util.TrimWhitespace(content) + "</li>"
		wroteItem = true
	}

	if !wroteItem {
		return out
	}
	return "<" + tag + formatHookAttributes(attrs) + ">" + out + "</" + tag + ">"
}
