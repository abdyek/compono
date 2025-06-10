package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
)

type Renderer interface {
	Render(writer io.Writer, source []byte, root ast.Node) error
}

func DefaultRenderer() Renderer {
	return &renderer{}
}

type renderer struct {
}

func (r *renderer) Render(writer io.Writer, source []byte, root ast.Node) error {
	return nil
}
