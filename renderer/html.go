package renderer

import (
	"io"
	"regexp"

	gohtml "html"

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
		html.NewP(),
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

		if child.Rule().Name() == "plain" {
			result += hr.prepareRaw(child)
		}

		el, ok := hr.elementMap[child.Rule().Name()]
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

func (hr *htmlRenderer) prepareRaw(node ast.Node) string {

	raw := string(node.Raw())
	escaped := gohtml.EscapeString(raw)

	trimmed := hr.dynamicTrim(node, escaped)

	return trimmed
}

func (hr *htmlRenderer) dynamicTrim(node ast.Node, raw string) string {
	// TODO: Complete it. There are some exceptions
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(raw, " ")
}

type element interface {
	Name() string
	Void() bool
}
