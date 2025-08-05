package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
)

type Renderer interface {
	Render(writer io.Writer, root ast.Node) error
}

func DefaultRenderer() Renderer {
	return newHtmlRenderer()
}

type renderer struct {
}

func (r *renderer) Render(writer io.Writer, root ast.Node) error {

	_, err := writer.Write([]byte(r.render(root.Children())))
	if err != nil {
		return err
	}

	return nil
}

func (r *renderer) render(children []ast.Node) string {

	if len(children) == 0 {
		return ""
	}

	result := ""

	for _, child := range children {
		if cn := child.Component().Name(); cn == "h1-content" {
			result += "<h1>" + r.render(child.Children()) + "</h1>"
		} else if cn == "p-content" {
			result += "<p>" + r.render(child.Children()) + "</p>"
		} else if cn == "h2-content" {
			result += "<h2>" + r.render(child.Children()) + "</h2>"
		} else if cn == "plain" {
			result += string(child.Raw())
		} else {
			result += r.render(child.Children())
		}
	}

	return result
}
