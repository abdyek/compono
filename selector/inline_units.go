package selector

import "bytes"

// InlineUnits returns the code spans and the {{ }} units of source, recognized
// from left to right: whichever opens first wins, and the delimiters of the
// other inside it are plain text.
func InlineUnits(source []byte) (codes [][2]int, units [][2]int) {
	codes, units, _, _ = scanInlineUnits(source)
	return codes, units
}

// scanInlineUnits recognizes code spans and {{ }} units from left to right.
// units holds every closed unit, invalid the closed units that follow no
// grammar, and unclosed every {{ that is not closed.
func scanInlineUnits(source []byte) (codes, units, invalid, unclosed [][2]int) {
	i := 0
	for i < len(source) {
		switch source[i] {
		case '`':
			n := 0
			for i+n < len(source) && source[i+n] == '`' {
				n++
			}
			if n == 1 {
				start := i
				j := i + 1
				matched := false
				for j < len(source) && source[j] != '\n' {
					if source[j] == '`' {
						runLen := 0
						for j+runLen < len(source) && source[j+runLen] == '`' {
							runLen++
						}
						if runLen == 1 {
							codes = append(codes, [2]int{start, j + 1})
							i = j + 1
							matched = true
						}
						break
					}
					j++
				}
				if !matched {
					i = start + n
				}
				continue
			}
			i += n
		case '{':
			if i+1 < len(source) && source[i+1] == '{' {
				end, valid, closed := unitEnd(source, i)
				if !closed {
					unclosed = append(unclosed, [2]int{i, i + 2})
					i += 2
					continue
				}
				units = append(units, [2]int{i, end})
				if !valid {
					invalid = append(invalid, [2]int{i, end})
				}
				i = end
				continue
			}
			i++
		default:
			i++
		}
	}
	return
}

// unitEnd returns where the {{ }} unit that starts at start ends. A unit that
// follows the grammar of a component call, a parameter reference or a context
// reference ends where its grammar ends, and valid is true. Otherwise it ends
// at the first }} after start. closed is false when there is no }} after start
// or another {{ comes before the first one.
func unitEnd(source []byte, start int) (end int, valid, closed bool) {
	if end, ok := scanMustacheCall(source, start, false); ok {
		return end, true, true
	}
	if end, ok := scanMustacheCall(source, start, true); ok {
		return end, true, true
	}
	if end, ok := scanContextMustacheCall(source, start); ok {
		return end, true, true
	}
	rest := source[start+2:]
	closeIdx := bytes.Index(rest, []byte("}}"))
	if closeIdx == -1 || bytes.Contains(rest[:closeIdx], []byte("{{")) {
		return 0, false, false
	}
	return start + 2 + closeIdx + 2, false, true
}

// ExpressionUnits returns the {{ }} units of source that follow no grammar
// and the {{ that are not closed, recognized like InlineUnits: a {{ inside a
// code span is text. Each unclosed range is the two bytes of its {{.
func ExpressionUnits(source []byte) (invalid, unclosed [][2]int) {
	_, _, invalid, unclosed = scanInlineUnits(source)
	return invalid, unclosed
}
