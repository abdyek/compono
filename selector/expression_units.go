package selector

type invalidExpressions struct {
	block bool
}

// NewInvalidExpressions selects the {{ }} units that follow no grammar. A
// unit's }} is searched for within the range it starts in. With block, a
// range is first split at every blank line, and only a unit that stands on
// its own lines is selected: nothing but spaces and tabs comes before it on
// its first line and after it on its last line.
func NewInvalidExpressions(block bool) Selector {
	return &invalidExpressions{block: block}
}

func (_ *invalidExpressions) Name() string {
	return "invalid_expressions"
}

func (s *invalidExpressions) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	for _, ns := range filterNoSelected(without, len(source)) {
		parts := [][2]int{ns}
		if s.block {
			parts = splitAtBlankLines(source, ns)
		}
		for _, part := range parts {
			invalid, _ := ExpressionUnits(source[part[0]:part[1]])
			for _, u := range invalid {
				u = [2]int{part[0] + u[0], part[0] + u[1]}
				if s.block && !standsOnOwnLines(source, u) {
					continue
				}
				results = append(results, u)
			}
		}
	}
	return results
}

type unclosedExpressions struct{}

// NewUnclosedExpressions selects the two bytes of every {{ that is not
// closed within the range it starts in.
func NewUnclosedExpressions() Selector {
	return &unclosedExpressions{}
}

func (_ *unclosedExpressions) Name() string {
	return "unclosed_expressions"
}

func (_ *unclosedExpressions) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	for _, ns := range filterNoSelected(without, len(source)) {
		_, unclosed := ExpressionUnits(source[ns[0]:ns[1]])
		for _, u := range unclosed {
			results = append(results, [2]int{ns[0] + u[0], ns[0] + u[1]})
		}
	}
	return results
}

// splitAtBlankLines splits r so that no part contains "\n\n". A part ends
// right after the first newline of a blank line.
func splitAtBlankLines(source []byte, r [2]int) [][2]int {
	parts := [][2]int{}
	start := r[0]
	for j := r[0]; j+1 < r[1]; j++ {
		if source[j] == '\n' && source[j+1] == '\n' {
			parts = append(parts, [2]int{start, j + 1})
			start = j + 1
		}
	}
	return append(parts, [2]int{start, r[1]})
}

// standsOnOwnLines reports whether only spaces and tabs come before u on its
// first line and after u on its last line.
func standsOnOwnLines(source []byte, u [2]int) bool {
	for i := u[0] - 1; i >= 0 && source[i] != '\n'; i-- {
		if source[i] != ' ' && source[i] != '\t' {
			return false
		}
	}
	for i := u[1]; i < len(source) && source[i] != '\n'; i++ {
		if source[i] != ' ' && source[i] != '\t' {
			return false
		}
	}
	return true
}
