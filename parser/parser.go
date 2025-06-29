package parser

import (
	"sort"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/component"
	"github.com/umono-cms/compono/util"
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
	rootNode := ast.DefaultNode()
	rootNode.SetComponent(&component.Root{})
	return p.parse(source, rootNode, comps)
}

func (p *parser) parse(source []byte, parentNode ast.Node, comps []component.Component) ast.Node {

	alreadySelected := [][2]int{}

	found := []foundComp{}

	for _, comp := range comps {

		if util.InSliceString(parentNode.Component().Name(), comp.DisallowParent()) {
			continue
		}

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
		children = append(children, nodeForm)
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
