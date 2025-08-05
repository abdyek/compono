package compono

import (
	"io"

	"github.com/umono-cms/compono/component"
	"github.com/umono-cms/compono/parser"
	"github.com/umono-cms/compono/renderer"
)

type Compono interface {
	Convert(source []byte, writer io.Writer) error
	Parser() parser.Parser
	SetParser(parser.Parser)
	Renderer() renderer.Renderer
	SetRenderer(renderer.Renderer)
	Components() []component.Component
	RegisterComponents(...component.Component)
	UnregisterComponent(name string)
}

func New() Compono {
	return &compono{
		parser:   parser.DefaultParser(),
		renderer: renderer.DefaultRenderer(),
	}
}

type compono struct {
	parser     parser.Parser
	renderer   renderer.Renderer
	components []component.Component
}

func (c *compono) Convert(source []byte, writer io.Writer) error {
	root := c.parser.Parse(source)
	return c.renderer.Render(writer, root)
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

func (c *compono) Components() []component.Component {
	return c.components
}

// TODO: redesign
func (c *compono) RegisterComponents(comps ...component.Component) {
	c.components = component.OverrideComponents(c.components, comps)
}

// TODO: redesign
func (c *compono) UnregisterComponent(name string) {
	i, _ := component.FindCompIndexByName(c.components, name)

	if i == -1 {
		return
	}

	c.components = append(c.components[:i], c.components[i+1:]...)
}
