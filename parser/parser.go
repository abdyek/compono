package parser

import (
	"sort"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
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
	rootNode.SetRule(&rule.Root{})
	return p.parse(source, rootNode, rootNode.Rule().Rules())
}

func (p *parser) parse(source []byte, parentNode ast.Node, rules []rule.Rule) ast.Node {

	alreadySelected := [][2]int{}

	found := []foundRule{}

	for _, rule := range rules {

		for _, slctr := range rule.Selectors() {

			sort.Slice(alreadySelected, func(i, j int) bool {
				return alreadySelected[i][0] < alreadySelected[j][0]
			})

			indexes := slctr.Select(source, alreadySelected...)

			for _, index := range indexes {
				found = append(found, foundRule{
					rule:  rule,
					start: index[0],
					end:   index[1],
				})
				alreadySelected = append(alreadySelected, index)
			}
		}
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].start < found[j].start
	})

	children := []ast.Node{}

	for _, f := range found {
		nodeForm := ast.DefaultNode()
		nodeForm.SetRule(f.rule)
		nodeForm.SetRaw(source[f.start:f.end])
		children = append(children, nodeForm)
	}

	for i := 0; i < len(children); i++ {
		children[i] = p.parse(children[i].Raw(), children[i], children[i].Rule().Rules())
	}

	parentNode.SetChildren(children)

	return parentNode
}

type foundRule struct {
	rule  rule.Rule
	start int
	end   int
}
