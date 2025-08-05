package renderer

import (
	"io"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/renderer/html"
)

type htmlRenderer struct {
	elementMap map[string]element
}

func newHtmlRenderer() Renderer {

	renderer := &htmlRenderer{}
	renderer.elementMap = make(map[string]element)

	elements := []element{
		html.NewH1(),
		html.NewH2(),
		html.NewStrong(),
		html.NewEm(),
	}

	for _, el := range elements {
		renderer.elementMap[el.Name()] = el
	}

	return renderer
}

func (hr *htmlRenderer) Render(writer io.Writer, root ast.Node) error {

	_, err := writer.Write([]byte(hr.render(root.Children())))
	if err != nil {
		return err
	}

	return nil
}

func (hr *htmlRenderer) render(children []ast.Node) string {

	if len(children) == 0 {
		return ""
	}

	result := ""

	for _, child := range children {

		if child.Component().Name() == "plain" {
			// TODO: Add html escape filter
			result += string(child.Raw())
		}

		el, ok := hr.elementMap[child.Component().Name()]
		if !ok {
			result += hr.render(child.Children())
			continue
		}

		if el.Void() {
			result += "<" + el.Name() + ">"
		} else {
			result += "<" + el.Name() + ">" + hr.render(child.Children()) + "</" + el.Name() + ">"
		}
	}

	return result
}

type element interface {
	Name() string
	Void() bool
}
