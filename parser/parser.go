package parser

import (
	"sort"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/component"
)

type Parser interface {
	Parse(source []byte) ast.Node
}

func DefaultParser() Parser {
	return &parser{}
}

type parser struct {
}

func (p *parser) Parse(source []byte) ast.Node {
	rootNode := ast.DefaultNode()
	rootNode.SetComponent(&component.Root{})
	return p.parse(source, rootNode, rootNode.Component().Components())
}

func (p *parser) parse(source []byte, parentNode ast.Node, comps []component.Component) ast.Node {

	alreadySelected := [][2]int{}

	found := []foundComp{}

	for _, comp := range comps {

		for _, slctr := range comp.Selectors() {
			indexes := slctr.Select(source, alreadySelected...)

			for _, index := range indexes {
				if freeToSelect(alreadySelected, index[0], index[1]) {
					found = append(found, foundComp{
						comp:  comp,
						start: index[0],
						end:   index[1],
					})
					alreadySelected = append(alreadySelected, index)
				}
			}
		}
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].start < found[j].start
	})

	children := []ast.Node{}

	for _, f := range found {
		nodeForm := ast.DefaultNode()
		nodeForm.SetComponent(f.comp)
		nodeForm.SetRaw(source[f.start:f.end])
		children = append(children, nodeForm)
	}

	for i := 0; i < len(children); i++ {
		children[i] = p.parse(children[i].Raw(), children[i], children[i].Component().Components())
	}

	parentNode.SetChildren(children)

	return parentNode
}

type foundComp struct {
	comp  component.Component
	start int
	end   int
}

func freeToSelect(alreadySelected [][2]int, start, end int) bool {
	for _, selected := range alreadySelected {
		if (start >= selected[0] && start < selected[1]) || (end > selected[0] && end <= selected[1]) {
			return false
		}
	}
	return true
}
