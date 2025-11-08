package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/selector"
	"github.com/umono-cms/compono/testdata/mocks"
)

type parserTestSuite struct {
	suite.Suite
}

func (s *parserTestSuite) TestParse() {
	for _, tt := range []struct {
		// TODO: Improve it
		source string
		node   ast.Node
		result ast.Node
	}{
		{
			source: `AAABBBCCC`,
			node: mocks.NewNodeBuilder().WithRule(
				mocks.NewRuleBuilder().WithName("abc-wrapper").WithSelectors([]selector.Selector{
					mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
						return [][2]int{[2]int{0, 9}}
					}),
				}).WithRules([]rule.Rule{
					mocks.NewRuleBuilder().WithName("a").WithSelectors([]selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{0, 3}}
						}),
					}).Build(),
					mocks.NewRuleBuilder().WithName("b").WithSelectors([]selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{3, 6}}
						}),
					}).Build(),
					mocks.NewRuleBuilder().WithName("c").WithSelectors([]selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{6, 9}}
						}),
					}).Build(),
				}).Build(),
			).Build(),
			result: mocks.NewNodeBuilder().WithChildren(
				[]ast.Node{
					mocks.NewNodeBuilder().WithRule(
						mocks.NewRuleBuilder().WithName("a").WithSelectors([]selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{0, 3}}
							}),
						}).Build(),
					).Build(),
					mocks.NewNodeBuilder().WithRule(
						mocks.NewRuleBuilder().WithName("b").WithSelectors([]selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{3, 6}}
							}),
						}).Build(),
					).Build(),
					mocks.NewNodeBuilder().WithRule(
						mocks.NewRuleBuilder().WithName("c").WithSelectors([]selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{6, 9}}
							}),
						}).Build(),
					).Build(),
				},
			).WithRaw([]byte(`AAABBBCCC`)).Build(),
		},
		// TODO: increase test cases
	} {
		// TODO: Improve it
		p := DefaultParser(logger.NewLogger())
		resNode := p.Parse([]byte(tt.source), tt.node)
		assert.Equal(s.T(), 3, len(resNode.Children()))
	}
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(parserTestSuite))
}
