package parser

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/internal/testutil/mocks"
	"github.com/umono-cms/compono/logger"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/selector"
)

type parserTestSuite struct {
	suite.Suite
}

func (s *parserTestSuite) TestParse() {
	for _, tt := range []struct {
		name   string
		source string
		node   ast.Node
		tree   tree
	}{
		{
			name:   "Simple",
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
			tree: tree{
				ruleName: "abc-wrapper",
				raw:      `AAABBBCCC`,
				children: []tree{
					{
						ruleName:       "a",
						parentRuleName: "abc-wrapper",
						raw:            "AAA",
					},
					{
						ruleName:       "b",
						parentRuleName: "abc-wrapper",
						raw:            "BBB",
					},
					{
						ruleName:       "c",
						parentRuleName: "abc-wrapper",
						raw:            "CCC",
					},
				},
			},
		},
		{
			name:   "Nested",
			source: `ABAACDCA`,
			node: mocks.NewNodeBuilder().WithRule(
				mocks.NewRuleBuilder().WithName("wrapper").WithSelectors([]selector.Selector{
					mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
						return [][2]int{[2]int{0, 8}}
					}),
				}).WithRules([]rule.Rule{
					mocks.NewRuleBuilder().WithName("ab").WithSelectors([]selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{0, 3}}
						}),
					}).WithRules([]rule.Rule{
						mocks.NewRuleBuilder().WithName("b").WithSelectors([]selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{1, 2}}
							}),
						}).Build(),
					}).Build(),
					mocks.NewRuleBuilder().WithName("acd").WithSelectors([]selector.Selector{
						mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
							return [][2]int{[2]int{3, 8}}
						}),
					}).WithRules([]rule.Rule{
						mocks.NewRuleBuilder().WithName("cd").WithSelectors([]selector.Selector{
							mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
								return [][2]int{[2]int{1, 4}}
							}),
						}).WithRules([]rule.Rule{
							mocks.NewRuleBuilder().WithName("d").WithSelectors([]selector.Selector{
								mocks.NewSelector(func([]byte, ...[2]int) [][2]int {
									return [][2]int{[2]int{1, 2}}
								}),
							}).Build(),
						}).Build(),
					}).Build(),
				}).Build(),
			).Build(),
			tree: tree{
				ruleName: "wrapper",
				raw:      `ABAACDCA`,
				children: []tree{
					{
						ruleName:       "ab",
						parentRuleName: "wrapper",
						raw:            "ABA",
						children: []tree{
							{
								ruleName:       "b",
								parentRuleName: "ab",
								raw:            "B",
							},
						},
					},
					{
						ruleName:       "acd",
						parentRuleName: "wrapper",
						raw:            "ACDCA",
						children: []tree{
							{
								ruleName:       "cd",
								parentRuleName: "acd",
								raw:            "CDC",
								children: []tree{
									{
										ruleName:       "d",
										parentRuleName: "cd",
										raw:            "D",
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		p := DefaultParser(logger.NewLogger())
		result := p.Parse([]byte(tt.source), tt.node)
		tt.tree.compareWithResult(s, tt.name, result)
	}
}

func TestParserTestSuite(t *testing.T) {
	suite.Run(t, new(parserTestSuite))
}

func TestParseSetsRanges(t *testing.T) {
	p := DefaultParser(logger.NewLogger())

	root := p.Parse([]byte("a\n\n{{ FOO }}"), ast.DefaultRootNode())

	assert.Equal(t, ast.Range{Start: 0, End: 12}, root.Range())

	blockCompCalls := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleName(node, "block-comp-call")
	})
	require.Len(t, blockCompCalls, 1)

	blockCompCall := blockCompCalls[0]
	assert.Equal(t, ast.Range{Start: 3, End: 12}, blockCompCall.Range())
	assert.Equal(t, []byte("{{ FOO }}"), blockCompCall.Raw())
}

func TestParseRangesMatchRaw(t *testing.T) {
	p := DefaultParser(logger.NewLogger())

	cutRemoved := func(source []byte, r ast.Range, removed [][2]int) []byte {
		out := []byte{}
		pos := r.Start
		for _, rem := range removed {
			if rem[0] < r.Start || rem[1] > r.End {
				continue
			}
			out = append(out, source[pos:rem[0]]...)
			pos = rem[1]
		}
		out = append(out, source[pos:r.End]...)
		return out
	}

	assertRanges := func(t *testing.T, path string, source []byte, node ast.Node) {
		lineRemover, ok := node.Rule().(rule.LineRemover)
		require.True(t, ok, "rule %q must implement rule.LineRemover", node.Rule().Name())
		removed := lineRemover.RemovedLines(source)

		var walk func(ast.Node)
		walk = func(n ast.Node) {
			r := n.Range()
			assert.Equal(t, string(cutRemoved(source, r, removed)), string(n.Raw()), "at %q", path)
			for _, child := range n.Children() {
				walk(child)
			}
		}
		walk(node)
	}

	rootFiles, err := filepath.Glob("../testdata/*/input.comp")
	require.NoError(t, err)
	for _, path := range rootFiles {
		source, err := os.ReadFile(path)
		require.NoError(t, err)

		node := ast.DefaultEmptyNode()
		node.SetRule(rule.NewRoot())

		node = p.Parse(source, node)
		assertRanges(t, path, source, node)
	}

	globalDirs, err := filepath.Glob("../testdata/*/global")
	require.NoError(t, err)
	for _, dir := range globalDirs {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
			require.NoError(t, walkErr)
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".comp" {
				return nil
			}
			source, err := os.ReadFile(path)
			require.NoError(t, err)

			node := ast.DefaultEmptyNode()
			node.SetRule(rule.NewGlobalCompDef())

			node = p.Parse(source, node)
			assertRanges(t, path, source, node)
			return nil
		})
		require.NoError(t, err)
	}
}

func TestParseCommentLineRanges(t *testing.T) {
	p := DefaultParser(logger.NewLogger())

	root := p.Parse([]byte("a\n// x\nb"), ast.DefaultRootNode())

	assert.Equal(t, []byte("a\nb"), root.Raw())
	assert.Equal(t, ast.Range{Start: 0, End: 8}, root.Range())

	ps := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleName(node, "p")
	})
	require.Len(t, ps, 1)
	assert.Equal(t, ast.Range{Start: 0, End: 8}, ps[0].Range())

	plains := ast.FilterNodesInTree(root, func(node ast.Node) bool {
		return ast.IsRuleName(node, "plain") && string(node.Raw()) == "b"
	})
	require.Len(t, plains, 1)
	assert.Equal(t, ast.Range{Start: 7, End: 8}, plains[0].Range())
}

type tree struct {
	ruleName       string
	parentRuleName string
	raw            string
	children       []tree
}

func (t tree) compareWithResult(s suite.TestingSuite, name string, n ast.Node) {
	assert.Equal(s.T(), t.ruleName, n.Rule().Name(), "at %q", name)
	if t.parentRuleName != "" {
		require.NotNil(s.T(), n.Parent())
		require.NotNil(s.T(), n.Parent().Rule())
		assert.Equal(s.T(), t.parentRuleName, n.Parent().Rule().Name())
	}
	assert.Equal(s.T(), []byte(t.raw), n.Raw(), "at %q", name)
	assert.Equal(s.T(), len(t.children), len(n.Children()), "at %q", name)
	for i, tc := range t.children {
		tc.compareWithResult(s, name, n.Children()[i])
	}
}
