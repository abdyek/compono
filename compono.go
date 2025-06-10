package compono

import (
	"io"

	"github.com/umono-cms/compono/components"
	"github.com/umono-cms/compono/parser"
	"github.com/umono-cms/compono/renderer"
)

type Compono interface {
	Convert(source []byte, writer io.Writer) error
	Parser() parser.Parser
	SetParser(parser.Parser)
	Renderer() renderer.Renderer
	SetRenderer(renderer.Renderer)
}

func New() Compono {
	return &compono{
		parser:   parser.DefaultParser(),
		renderer: renderer.DefaultRenderer(),
	}
}

type compono struct {
	parser   parser.Parser
	renderer renderer.Renderer
}

func (c *compono) Convert(source []byte, writer io.Writer) error {

	root := c.parser.Parse(source, []components.Component{})

	return c.renderer.Render(writer, source, root)
}

func (c *compono) Parser() parser.Parser {
	return c.parser
}

func (c *compono) SetParser(parser parser.Parser) {
	c.parser = parser
}

func (c *compono) Renderer() renderer.Renderer {
	return c.renderer
}

func (c *compono) SetRenderer(renderer renderer.Renderer) {
	c.renderer = renderer
}
