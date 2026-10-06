package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/renderer/html"
)

type Renderer interface {
	Render(writer io.Writer, root ast.Node) error
}

type ErrorStylesheetSetter interface {
	SetErrorStylesheet(string)
}

func DefaultRenderer(log logger.Logger) Renderer {
	return html.NewRenderer(log)
}
