package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/renderer/html"
)

// Options holds the settings of a single Render call.
type Options = html.Options

// Diagnostic is a part of the output that an error dropped.
type Diagnostic = html.Diagnostic

// Call is a component call on the render stack.
type Call = html.Call

type Renderer interface {
	Render(writer io.Writer, root ast.Node, opts Options) ([]Diagnostic, error)
}

func DefaultRenderer(log logger.Logger) Renderer {
	return html.NewRenderer(log)
}
