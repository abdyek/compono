package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type startEndTestSuite struct {
	suite.Suite
}

func (s *startEndTestSuite) TestName() {
	se := &startEnd{}
	require.Equal(s.T(), "start_end", se.Name())
}

func (s *startEndTestSuite) TestSelect() {
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
			source:    "xxx(yyy)zzz",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{3, 8}},
		},
		{
			name:      "No match",
			source:    "xxxyyyzzz",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{},
		},
		{
			name:      "Regular siblings",
			source:    "abc(no-matter)xyz(another-no-matter)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{17, 36}, {3, 14}},
		},
		{
			name:      "Immediately adjacent siblings",
			source:    "abc(no-matter)(another-no-matter)xyz",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{14, 33}, {3, 14}},
		},
		{
			name:      "Filled",
			source:    "(abc)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{0, 5}},
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
			selected:  [][2]int{{0, 2}},
		},
		{
			name:      "Nested",
			source:    "(A (B) C)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{0, 9}},
		},
		{
			name:      "Irregular 1",
			source:    "abcde((123)xyz",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{6, 11}},
		},
		{
			name:      "Irregular 2",
			source:    "abcde(123))xyz",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{5, 10}},
		},
		{
			name:      "Irregular 3",
			source:    "(abc)de)f(g(h)ij",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{11, 14}, {0, 5}},
		},
		{
			name:      "Regular at first",
			source:    "(hello)!!",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{0, 7}},
		},
		{
			name:      "Regular at last",
			source:    "!!!(hello)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{3, 10}},
		},
		{
			name:      "Regular at last",
			source:    "!!!(hello)",
			startWith: `\(`,
			endWith:   `\)`,
			without:   nil,
			selected:  [][2]int{{3, 10}},
		},
		{
			name:      "At least 2 letters regex - Regular",
			source:    "abc{{ COMP )))xyz",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{{3, 14}},
		},
		{
			name:      "At least 2 letters regex - No match",
			source:    "abc COMP xyz",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{},
		},
		{
			name:      "At least 2 letters regex - Regular siblings",
			source:    "abc {{ XYZ ))) {{ ABC ))) xyz",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{{15, 25}, {4, 14}},
		},
		{
			name:      "At least 2 letters regex - Immediately adjacent siblings",
			source:    "abc {{ XYZ ))){{ ABC ))) xyz",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{{14, 24}, {4, 14}},
		},
		{
			name:      "At least 2 letters regex - Nested",
			source:    "abc {{ xyz {{ nested ))) xyz)))",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{{4, 31}},
		},
		{
			name:      "At least 2 letters regex - Irregular 1",
			source:    " {{abc{{ abc )))",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{{6, 16}},
		},
		{
			name:      "At least 2 letters regex - Irregular 2",
			source:    "abc{{ abc ))) )))",
			startWith: `\{\{`,
			endWith:   `\)\)\)`,
			without:   nil,
			selected:  [][2]int{{3, 13}},
		},
		{
			name:      "With metacharacters",
			source:    "abcde",
			startWith: `^`,
			endWith:   `$`,
			without:   nil,
			selected:  [][2]int{{0, 5}},
		},
		{
			name:      "UTF-8 regexp",
			source:    "aböçcdşıe",
			startWith: `öç`,
			endWith:   `şı`,
			without:   nil,
			selected:  [][2]int{{2, 12}},
		},
		{
			name:      "Same regexp",
			source:    "abc**cde**",
			startWith: `\*\*`,
			endWith:   `\*\*`,
			without:   nil,
			selected:  [][2]int{{3, 10}},
		},
		{
			name:      "Start regex substring of end regexp",
			source:    "abc foo xyz foobar 012",
			startWith: `foo`,
			endWith:   `foobar`,
			without:   nil,
			selected:  [][2]int{{4, 18}},
		},
		{
			name:      "End regex substring of start regexp",
			source:    "abc foobar xyz foo 012",
			startWith: `foobar`,
			endWith:   `foo`,
			without:   nil,
			selected:  [][2]int{{4, 18}},
		},
		{
			name:      "With single without",
			source:    "abcde without {{ COMP }} xyz",
			startWith: `\{\{`,
			endWith:   `\}\}`,
			without:   [][2]int{{6, 13}},
			selected:  [][2]int{{14, 24}},
		},
		{
			name:      "With multi withouts",
			source:    "abcde without {{ COMP }} without xyz",
			startWith: `\{\{`,
			endWith:   `\}\}`,
			without:   [][2]int{{6, 13}, {25, 32}},
			selected:  [][2]int{{14, 24}},
		},
		{
			name:      "With filled without",
			source:    "abcde without {{ COMP }} without xyz",
			startWith: `\{\{`,
			endWith:   `\}\}`,
			without:   [][2]int{{0, 13}, {25, 36}},
			selected:  [][2]int{{14, 24}},
		},
	} {
		se, err := NewStartEnd(tt.startWith, tt.endWith)
		require.Nil(s.T(), err, "at '"+tt.name+"'")
		selected := se.Select([]byte(tt.source), tt.without...)
		assert.Equal(s.T(), tt.selected, selected, "at '"+tt.name+"'")
	}
}

func TestStartEndTestSuite(t *testing.T) {
	suite.Run(t, new(startEndTestSuite))
}
