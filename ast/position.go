package ast

import "unicode/utf8"

// Position is a point in a source: a 0-based byte offset, a 1-based line and a
// 1-based column counted in runes.
type Position struct {
	Offset int
	Line   int
	Column int
}

// PositionAt returns the position of the byte offset in source. Lines are
// separated by \n.
func PositionAt(source []byte, offset int) Position {
	line := 1
	lastNewline := -1
	for i := 0; i < offset; i++ {
		if source[i] == '\n' {
			line++
			lastNewline = i
		}
	}
	column := 1 + utf8.RuneCount(source[lastNewline+1:offset])
	return Position{Offset: offset, Line: line, Column: column}
}
