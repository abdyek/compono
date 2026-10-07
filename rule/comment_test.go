package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommentLines(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   [][2]int
	}{
		{"simple", "// a\nb", [][2]int{{0, 5}}},
		{"indented spaces", "  // a\nb", [][2]int{{0, 7}}},
		{"tab", "\t//\nb", [][2]int{{0, 4}}},
		{"triple slash", "/// a", [][2]int{{0, 5}}},
		{"trailing comment", "a // b\nc", [][2]int{}},
		{"hash before comment", "# // a", [][2]int{}},
		{"second line", "a\n// b", [][2]int{{2, 6}}},
		{"consecutive comments", "a\n// b\n// c\nd", [][2]int{{2, 7}, {7, 12}}},
		{"code fence", "```\n// a\n```\n// b\n", [][2]int{{13, 18}}},
		{"closed unit", "{{ X\n// a\n}}\nb", [][2]int{}},
		{"unclosed unit", "{{ X\n// a\nb", [][2]int{{5, 10}}},
		{"nested open unit", "{{ X\n// a\n{{ Y }}", [][2]int{{5, 10}}},
		{"comment opens nothing", "// {{ X\nb }}", [][2]int{{0, 8}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, commentLines([]byte(tt.source)))
		})
	}
}
