package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type startEndLeftInnerTestSuite struct {
	suite.Suite
}

func (s *startEndLeftInnerTestSuite) TestName() {
	seli := &startEndLeftInner{}
	require.Equal(s.T(), "start_end_left_inner", seli.Name())
}

func (s *startEndLeftInnerTestSuite) TestSelect() {
	for _, tt := range []struct {
		name      string
		source    string
		startWith string
		endWith   string
		without   [][2]int
		selected  [][2]int
	}{
		{
			name:      "Regular",
			source:    "{ COMP_NAME }",
			startWith: `\{`,
			endWith:   `\}`,
			without:   nil,
			selected:  [][2]int{{0, 12}},
		},
		{
			name:      "No match",
			source:    "no match",
			startWith: `\{`,
			endWith:   `\}`,
			without:   nil,
			selected:  [][2]int{},
		},
		{
			name:      "Regular siblings",
			source:    "abc{ COMP_NAME }xyz{ COMP_NAME_2 }012",
			startWith: `\{`,
			endWith:   `\}`,
			without:   nil,
			selected:  [][2]int{{19, 33}, {3, 15}},
		},
		{
			name:      "Immediately adjacent siblings",
			source:    "abc{ COMP_NAME }{ COMP_NAME_2 }012",
			startWith: `\{`,
			endWith:   `\}`,
			without:   nil,
			selected:  [][2]int{{16, 30}, {3, 15}},
		},
		{
			name:      "Filled",
			source:    "(abc)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{0, 4}},
		},
		{
			name:      "Empty content",
			source:    "",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{},
		},
		{
			name:      "Empty",
			source:    "()",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{0, 1}},
		},
		{
			name:      "Nested",
			source:    "(A (B) C)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{0, 8}},
		},
		{
			name:      "Irregular 1",
			source:    "{{abc}}de}}f{{g{{h}}ij",
			startWith: `\{\{`,
			endWith:   `\}\}`,
			without:   nil,
			selected:  [][2]int{{15, 18}, {0, 5}},
		},
		{
			name:      "No start",
			source:    "abcde }} xyz",
			startWith: `\{\{`,
			endWith:   `\}\}`,
			without:   nil,
			selected:  [][2]int{},
		},
		{
			name:      "Without",
			source:    "abcd {{ COMP }} without {{ COMP_2 }}",
			startWith: `\{\{`,
			endWith:   `\}\}`,
			without:   [][2]int{{16, 23}},
			selected:  [][2]int{{5, 13}, {24, 34}},
		},
	} {
		seli, err := NewStartEndLeftInner(tt.startWith, tt.endWith)
		require.Nil(s.T(), err, "at '"+tt.name+"'")
		selected := seli.Select([]byte(tt.source), tt.without...)
		assert.Equal(s.T(), tt.selected, selected, "at '"+tt.name+"'")
	}
}

func TestStartEndLeftInnerTestSuite(t *testing.T) {
	suite.Run(t, new(startEndLeftInnerTestSuite))
}
