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
			node: mocks.NewNode(
				mocks.NewRule("abc-wrapper",
					[]selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{0, 9}}
						}),
					},
					[]rule.Rule{
						mocks.NewRule("a", []selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{0, 3}}
							}),
						}, nil),
						mocks.NewRule("b", []selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{3, 6}}
							}),
						}, nil),
						mocks.NewRule("c", []selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{6, 9}}
							}),
						}, nil),
					}), nil, nil, nil,
			),
			result: mocks.NewNode(nil, nil, []ast.Node{
				mocks.NewNode(
					mocks.NewRule("a", []selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{0, 3}}
						}),
					}, nil), nil, nil, nil,
				),
				mocks.NewNode(
					mocks.NewRule("b", []selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{3, 6}}
						}),
					}, nil), nil, nil, nil,
				),
				mocks.NewNode(
					mocks.NewRule("c", []selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{6, 9}}
						}),
					}, nil), nil, nil, nil,
				),
			}, []byte(`AAABBBCCC`)),
		},
		// TODO: increase test cases
	} {
		log := logger.NewLogger()
		p := DefaultParser(log)
		resNode := p.Parse([]byte(tt.source), tt.node)
		assert.Equal(s.T(), 3, len(resNode.Children()))
	}
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(parserTestSuite))
}
