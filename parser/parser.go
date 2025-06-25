package parser

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/component"
)

type Parser interface {
	Parse(source []byte, comps []component.Component) ast.Node
}

func DefaultParser() Parser {
	return &parser{}
}

type parser struct {
}

func (p *parser) Parse(source []byte, comps []component.Component) ast.Node {
	return nil
}
