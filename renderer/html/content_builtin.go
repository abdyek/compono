package html

import (
	"fmt"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/internal/attrhook"
	"github.com/umono-cms/compono/util"
)

// contentBuiltin is a built-in that wraps its content component in a single
// element, such as MAIN in <main>.
type contentBuiltin struct {
	renderer *renderer
	name     string
	tag      string
	label    bool
	details  bool
}

func newContentBuiltin(rend *renderer, name, tag string, label, details bool) builtinComponent {
	return &contentBuiltin{
		renderer: rend,
		name:     name,
		tag:      tag,
		label:    label,
		details:  details,
	}
}

// New returns a new instance of the built-in bound to the same renderer.
func (cb *contentBuiltin) New() builtinComponent {
	return newContentBuiltin(cb.renderer, cb.name, cb.tag, cb.label, cb.details)
}

// Name returns the name of the built-in component.
func (cb *contentBuiltin) Name() string {
	return cb.name
}

// Render renders the built-in, wrapping the component value given by the
// "content" parameter in the tag of the built-in.
func (cb *contentBuiltin) Render(invoker renderableNode, node ast.Node) string {
	r := cb.renderer

	if title, message, ok := r.contentBuiltinPlacementError(cb.name, node); ok {
		return r.recordDiagnostic(node, title, message, true)
	}

	value := resolveBuiltinParam(r, invoker, node, "content")
	if value.Type != "comp" || value.Raw == "" {
		return ""
	}

	if ast.FindCompDefInScope(r.root, value.Scope, node, value.Raw) == nil {
		return r.recordDiagnostic(node, "Unknown component", "The component **"+value.Raw+"** is not defined.", true)
	}

	conflicts := map[string]bool{}
	labelAttr := ""
	if cb.label {
		conflicts = map[string]bool{"aria-label": true}
		if label, ok := getArgValue(r, invoker, node, "label"); ok && label != "" {
			labelAttr = " aria-label=\"" + label + "\""
		}
	}

	if cb.details {
		conflicts = map[string]bool{"open": true}
	}

	attrs := r.callAttributeHook(cb.name, conflicts)

	r.pushBuiltinFrame(cb.name, node)
	content := renderCompValue(r, newPassthroughRenderable(ast.NewCompValueFrame(value, node), invoker), value, false)
	r.popFrame()

	if isEmptyRendered(content) {
		return content
	}

	if cb.details {
		openAttr := ""
		if open, ok := getBoolArgValue(r, invoker, node, "open"); ok && open {
			openAttr = " open"
		}
		text := getArgValueWithDefa(r, invoker, node, "text", "")
		summary := "<summary>" + text + "</summary>"
		return "<details" + openAttr + formatHookAttributes(attrs) + ">" + summary + content + "</details>"
	}

	return "<" + cb.tag + labelAttr + formatHookAttributes(attrs) + ">" + content + "</" + cb.tag + ">"
}

// contentBuiltinPlacementError reports whether the built-in with name is placed
// in a forbidden frame. The look-up walks the frame stack from nearest (last)
// outwards.
func (r *renderer) contentBuiltinPlacementError(name string, node ast.Node) (title, message string, ok bool) {
	if name == "MAIN" {
		for i := len(r.frameStack) - 1; i >= 0; i-- {
			f := r.frameStack[i]
			if f.Kind != attrhook.FrameBuiltin {
				continue
			}
			return "Invalid component usage", "The built-in component **MAIN** cannot be placed inside **" + f.Name + "**.", true
		}
		if r.firstMainCall != "" && r.firstMainCall != r.callOccurrence(node) {
			return "Invalid component usage", "The built-in component **MAIN** can be used only once in a conversion.", true
		}
		return "", "", false
	}

	if name == "HEADER" || name == "FOOTER" {
		for i := len(r.frameStack) - 1; i >= 0; i-- {
			f := r.frameStack[i]
			if f.Kind != attrhook.FrameBuiltin {
				continue
			}
			if f.Name != "HEADER" && f.Name != "FOOTER" {
				continue
			}
			return "Invalid component usage", "The built-in component **" + name + "** cannot be placed inside **" + f.Name + "**.", true
		}
		return "", "", false
	}

	return "", "", false
}

// callOccurrence identifies a call in the current render: the call node and
// the call nodes around it.
func (r *renderer) callOccurrence(node ast.Node) string {
	result := fmt.Sprintf("%p", node)
	for _, n := range r.callNodes {
		result += "/" + fmt.Sprintf("%p", n)
	}
	return result
}

// noteBuiltinCall records the first MAIN call of the conversion.
func (r *renderer) noteBuiltinCall(name string, node ast.Node) {
	if name == "MAIN" && r.firstMainCall == "" {
		r.firstMainCall = r.callOccurrence(node)
	}
}

// isEmptyRendered reports whether rendered output has no content once the
// diagnostic and URL report markers are removed.
func isEmptyRendered(rendered string) bool {
	rest := diagnosticMarkerPattern.ReplaceAllString(rendered, "")
	rest = urlReportMarkerPattern.ReplaceAllString(rest, "")
	return util.TrimWhitespace(rest) == ""
}
