package compono

import (
	"io"

	"github.com/umono-cms/compono/parser"
	"github.com/umono-cms/compono/renderer"
	"github.com/umono-cms/compono/rule"
)

type Compono interface {
	Convert(source []byte, writer io.Writer) error
	Parser() parser.Parser
	SetParser(parser.Parser)
	Renderer() renderer.Renderer
	SetRenderer(renderer.Renderer)
	Rules() []rule.Rule
	RegisterRules(...rule.Rule)
	UnregisterComponent(name string)
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
	rules    []rule.Rule
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

func (c *compono) Rules() []rule.Rule {
	return c.rules
}

// TODO: redesign
func (c *compono) RegisterRules(rules ...rule.Rule) {
	c.rules = rule.OverrideRules(c.rules, rules)
}

// TODO: redesign
func (c *compono) UnregisterComponent(name string) {
	i, _ := rule.FindRuleIndexByName(c.rules, name)

	if i == -1 {
		return
	}

	c.rules = append(c.rules[:i], c.rules[i+1:]...)
}
