package html

import "github.com/umono-cms/compono/ast"

// TODO: Complete it
type builtinComponent interface {
	Name() string
	Render(ast.Node) string
}

type link struct {
	renderer *renderer
}

func NewLink(rend *renderer) builtinComponent {
	return &link{
		renderer: rend,
	}
}

func (_ *link) Name() string {
	return "LINK"
}

func (_ *link) Render(node ast.Node) string {
	return "<a href='https://umono.io' target='_blank'>Umono</a>"
}
