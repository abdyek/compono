package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/renderer/html"
)

// Options holds the settings of a single Render call.
type Options = html.Options

type Renderer interface {
	Render(writer io.Writer, root ast.Node, opts Options) error
}

func DefaultRenderer(log logger.Logger) Renderer {
	return html.NewRenderer(log)
}
