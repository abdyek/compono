package parser

import (
	"sort"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/selector"
)

type Parser interface {
	Parse(source []byte, node ast.Node) ast.Node
}

func DefaultParser(log logger.Logger) Parser {
	return &parser{logger: log}
}

type parser struct {
	logger logger.Logger
}

func (p *parser) Parse(source []byte, root ast.Node) ast.Node {
	cp := &parser{logger: logger.Scoped(p.logger)}
	cp.logger.Enter(logger.Parser, "Parser started")
	root.SetRange(ast.Range{Start: 0, End: len(source)})

	parseSource := source
	var offsets []int
	if lineRemover, ok := root.Rule().(rule.LineRemover); ok {
		if removed := lineRemover.RemovedLines(source); len(removed) > 0 {
			parseSource, offsets = stripRemovedLines(source, removed)
		}
	}

	node := cp.parse(parseSource, root)

	if offsets != nil {
		remapRanges(node, offsets, len(source))
		root.SetRange(ast.Range{Start: 0, End: len(source)})
	}

	cp.logger.Exit(logger.Parser, "Parser finished")
	return node
}

// stripRemovedLines returns the source without the given removed [start, end)
// ranges and an offsets slice where offsets[i] is the original offset of byte i
// of the stripped source.
func stripRemovedLines(source []byte, removed [][2]int) ([]byte, []int) {
	ranges := make([][2]int, len(removed))
	copy(ranges, removed)
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i][0] < ranges[j][0]
	})

	stripped := make([]byte, 0, len(source))
	offsets := make([]int, 0, len(source))

	pos := 0
	for _, rem := range ranges {
		if rem[0] < pos || rem[1] > len(source) {
			continue
		}
		for i := pos; i < rem[0]; i++ {
			offsets = append(offsets, i)
			stripped = append(stripped, source[i])
		}
		pos = rem[1]
	}
	for i := pos; i < len(source); i++ {
		offsets = append(offsets, i)
		stripped = append(stripped, source[i])
	}

	return stripped, offsets
}

// remapRanges maps the ranges of every descendant of node from the stripped
// source back to the original source. node itself keeps its range.
func remapRanges(node ast.Node, offsets []int, sourceLen int) {
	for _, child := range node.Children() {
		remapRange(child, offsets, sourceLen)
		remapRanges(child, offsets, sourceLen)
	}
}

func remapRange(node ast.Node, offsets []int, sourceLen int) {
	rng := node.Range()
	start := sourceLen
	if rng.Start < len(offsets) {
		start = offsets[rng.Start]
	}
	end := start
	if rng.End > rng.Start {
		end = offsets[rng.End-1] + 1
	}
	node.SetRange(ast.Range{Start: start, End: end})
}

func (p *parser) parse(source []byte, parentNode ast.Node) ast.Node {

	parentNode.SetRaw(source)

	p.logger.Enter(logger.Parser, "Parser started for %s", parentNode.Rule().Name())

	alreadySelected := [][2]int{}

	found := []foundRule{}

	for _, rule := range parentNode.Rule().Rules() {

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

			// TODO: Move it into the logger msg function. Because Parser don't need ordered indexes.
			sort.Slice(indexes, func(i, j int) bool {
				return indexes[i][0] < indexes[j][0]
			})
			// ^^^^ here

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
		nodeForm := ast.DefaultEmptyNode()
		nodeForm.SetRule(f.rule)
		nodeForm.SetRaw(source[f.start:f.end])
		nodeForm.SetParent(parentNode)
		nodeForm.SetRange(ast.Range{Start: parentNode.Range().Start + f.start, End: parentNode.Range().Start + f.end})
		children = append(children, nodeForm)
	}

	for i := 0; i < len(children); i++ {
		children[i] = p.parse(children[i].Raw(), children[i])
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
