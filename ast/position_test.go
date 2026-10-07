package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositionAt(t *testing.T) {
	source := "ab\nçd\n"

	tests := []struct {
		name     string
		offset   int
		expected Position
	}{
		{name: "start of source", offset: 0, expected: Position{Offset: 0, Line: 1, Column: 1}},
		{name: "after ascii chars", offset: 2, expected: Position{Offset: 2, Line: 1, Column: 3}},
		{name: "start of second line", offset: 3, expected: Position{Offset: 3, Line: 2, Column: 1}},
		{name: "after multibyte rune", offset: 5, expected: Position{Offset: 5, Line: 2, Column: 2}},
		{name: "end of second line", offset: 6, expected: Position{Offset: 6, Line: 2, Column: 3}},
		{name: "start of third line", offset: 7, expected: Position{Offset: 7, Line: 3, Column: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, PositionAt([]byte(source), tt.offset))
		})
	}
}
