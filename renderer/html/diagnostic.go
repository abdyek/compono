package html

import (
	"regexp"
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/internal/attrhook"
)

// Diagnostic is an error element written to the output.
type Diagnostic struct {
	Title   string
	Message string
	// Node is the node the error element is rendered from.
	Node ast.Node
	// Calls are the component calls around Node, outermost first.
	Calls []Call
}

// Call is a component call on the render stack.
type Call struct {
	Name string
	Kind attrhook.FrameKind
	// Node is the call: a component call, or the parameter reference that
	// renders a component value.
	Node ast.Node
}

var diagnosticMarkerPattern = regexp.MustCompile("\x00compono-diagnostic-([0-9]+)\x00")

func (r *renderer) recordDiagnostic(node ast.Node, title, message string) string {
	var calls []Call
	if len(r.frameStack) > 0 {
		calls = make([]Call, len(r.frameStack))
		for i, f := range r.frameStack {
			calls[i] = Call{
				Name: f.Name,
				Kind: f.Kind,
				Node: r.callNodes[i],
			}
		}
	}

	r.diagnostics = append(r.diagnostics, Diagnostic{
		Title:   title,
		Message: message,
		Node:    node,
		Calls:   calls,
	})

	return "\x00compono-diagnostic-" + strconv.Itoa(len(r.diagnostics)-1) + "\x00"
}

func (r *renderer) takeDiagnostics(out string) (string, []Diagnostic) {
	var diagnostics []Diagnostic

	result := diagnosticMarkerPattern.ReplaceAllStringFunc(out, func(match string) string {
		submatches := diagnosticMarkerPattern.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		idx, err := strconv.Atoi(submatches[1])
		if err != nil || idx < 0 || idx >= len(r.diagnostics) {
			return match
		}
		diagnostics = append(diagnostics, r.diagnostics[idx])
		return ""
	})

	return result, diagnostics
}
