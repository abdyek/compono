package rule

import (
	"github.com/umono-cms/compono/selector"
)

type table struct{}

func newTable() Rule {
	return &table{}
}

func (_ *table) Name() string {
	return "table"
}

func (_ *table) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`(?m)^[ \t]*\|[^\n]*(?:\n[ \t]*\|[^\n]*)*`)
	return []selector.Selector{
		p,
	}
}

func (_ *table) Rules() []Rule {
	return []Rule{
		newTableRow(),
	}
}

type tableRow struct{}

func newTableRow() Rule {
	return &tableRow{}
}

func (_ *tableRow) Name() string {
	return "table-row"
}

func (_ *tableRow) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`(?m)^[^\n]+`)
	return []selector.Selector{
		p,
	}
}

func (_ *tableRow) Rules() []Rule {
	return []Rule{
		newTableCell(),
	}
}

type tableCell struct{}

func newTableCell() Rule {
	return &tableCell{}
}

func (_ *tableCell) Name() string {
	return "table-cell"
}

func (_ *tableCell) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			filtered := [][2]int{}
			for _, ind := range index {
				start, end := ind[0], ind[1]
				seps := TableRowSeparators(source[start:end])
				for k := 0; k+1 < len(seps); k++ {
					filtered = append(filtered, [2]int{start + seps[k], start + seps[k+1]})
				}
			}
			return filtered
		}),
	}
}

func (_ *tableCell) RemovedLines(source []byte) [][2]int {
	if len(source) > 0 && source[0] == '|' {
		return [][2]int{{0, 1}}
	}
	return [][2]int{}
}

func (_ *tableCell) Rules() []Rule {
	return []Rule{
		newLink(),
		newInlineCode(),
		newStrong(),
		newEm(),
		newInlineCompCall(),
		newContextRef(),
		newParamRef(),
		newPlain(),
	}
}

// TableRowSeparators returns the offsets of the '|' characters of a table row
// that separate cells; a '|' inside a code span or a {{ }} unit is not a
// separator.
func TableRowSeparators(row []byte) []int {
	seps := []int{}
	i := 0
	for i < len(row) {
		c := row[i]
		if c == '`' {
			n := 0
			for i+n < len(row) && row[i+n] == '`' {
				n++
			}
			found := -1
			for j := i + n; j < len(row); j++ {
				if row[j] == '`' {
					m := 0
					for j+m < len(row) && row[j+m] == '`' {
						m++
					}
					if m == n {
						found = j + m
						break
					}
					j += m - 1
				}
			}
			if found >= 0 {
				i = found
			} else {
				i += n
			}
			continue
		}
		if c == '{' && i+1 < len(row) && row[i+1] == '{' {
			found := -1
			for j := i + 2; j+1 < len(row); j++ {
				if row[j] == '}' && row[j+1] == '}' {
					found = j + 2
					break
				}
			}
			if found >= 0 {
				i = found
			} else {
				i += 2
			}
			continue
		}
		if c == '|' {
			seps = append(seps, i)
			i++
			continue
		}
		i++
	}
	return seps
}
