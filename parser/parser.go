package parser

import (
	"sort"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/selector"
)

type Parser interface {
	Parse(source []byte) ast.Node
}

func DefaultParser(log logger.Logger) Parser {
	return &parser{logger: log}
}

type parser struct {
	logger logger.Logger
}

func (p *parser) Parse(source []byte) ast.Node {
	p.logger.Enter(logger.Parser, "Parser started")
	rootNode := ast.DefaultNode()
	rootNode.SetRule(&rule.Root{})
	rootNode = p.parse(source, rootNode, rootNode.Rule().Rules())
	p.logger.Exit(logger.Parser, "Parser finished")
	return rootNode
}

func (p *parser) parse(source []byte, parentNode ast.Node, rules []rule.Rule) ast.Node {

	p.logger.Enter(logger.Parser, "Parser started for %s", parentNode.Rule().Name())

	alreadySelected := [][2]int{}

	found := []foundRule{}

	for _, rule := range rules {

		p.logger.Enter(logger.Parser|logger.Detail, "Started searching for selectors of %s rule", logger.Colorize(logger.Bold(rule.Name()), logger.Green))

		for _, slctr := range rule.Selectors() {

			slctrName := "unknown"
			if n, ok := slctr.(selector.Named); ok {
				slctrName = n.Name()
			}

			p.logger.Log(logger.Parser|logger.Detail, "Started searching for %s selector", logger.Colorize(logger.Bold(slctrName), logger.Green))

			sort.Slice(alreadySelected, func(i, j int) bool {
				return alreadySelected[i][0] < alreadySelected[j][0]
			})

			indexes := slctr.Select(source, alreadySelected...)

			if len(indexes) != 0 {
				p.logger.Log(logger.Parser|logger.Detail, "Found indexes %v", indexes)
				p.logger.LogMultiline(logger.Parser|logger.Detail, "Source:\n%s", logger.Highlight(source, indexes, func(s string) string {
					return logger.Colorize(s, logger.Yellow)
				}))
			} else {
				p.logger.Log(logger.Parser|logger.Detail, "No indexes found")
			}

			for _, index := range indexes {
				found = append(found, foundRule{
					rule:  rule,
					start: index[0],
					end:   index[1],
				})
				alreadySelected = append(alreadySelected, index)
			}
		}

		p.logger.Exit(logger.Parser|logger.Detail, "Finished searching for selectors of %s rule", logger.Colorize(logger.Bold(rule.Name()), logger.Green))
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

	p.logger.Exit(logger.Parser, "Parser finished parsing for %s", parentNode.Rule().Name())

	return parentNode
}

type foundRule struct {
	rule  rule.Rule
	start int
	end   int
}
