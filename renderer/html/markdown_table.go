package html

import (
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
)

type markdownTable struct {
	baseRenderable
	renderer *renderer
}

func newMarkdownTable(rend *renderer) renderableNode {
	return &markdownTable{
		renderer: rend,
	}
}

func (m *markdownTable) New() renderableNode {
	return newMarkdownTable(m.renderer)
}

func (_ *markdownTable) Condition(_ renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "table")
}

func (m *markdownTable) Render() string {
	rows := m.Node().Children()
	firstRowCellCount := -1

	for i, row := range rows {
		raw := row.Raw()
		seps := rule.TableRowSeparators(raw)

		firstNonWS := -1
		lastNonWS := -1
		for k := 0; k < len(raw); k++ {
			if raw[k] != ' ' && raw[k] != '\t' {
				if firstNonWS == -1 {
					firstNonWS = k
				}
				lastNonWS = k
			}
		}
		if len(seps) < 2 || firstNonWS != seps[0] || lastNonWS != seps[len(seps)-1] {
			return m.renderer.recordDiagnostic(m.Node(), "Invalid table row", "The table row at index **"+strconv.Itoa(i)+"** must start and end with **|**.", true)
		}

		cells := make([]string, 0, len(seps)-1)
		for k := 0; k+1 < len(seps); k++ {
			cells = append(cells, string(raw[seps[k]+1:seps[k+1]]))
		}
		cellCount := len(cells)

		if i == 1 {
			valid := true
			for _, cell := range cells {
				trimmed := util.TrimWhitespace(cell)
				if len(trimmed) < 3 {
					valid = false
					break
				}
				for _, c := range trimmed {
					if c != '-' {
						valid = false
						break
					}
				}
				if !valid {
					break
				}
			}
			if !valid {
				return m.invalidDelimiter()
			}
		}

		if i > 0 {
			if cellCount != firstRowCellCount {
				return m.renderer.recordDiagnostic(m.Node(), "Table cell count mismatch", "The table row at index **"+strconv.Itoa(i)+"** has **"+strconv.Itoa(cellCount)+"** cells, but the header row has **"+strconv.Itoa(firstRowCellCount)+"**.", true)
			}
		}

		if i == 0 {
			for j, cell := range cells {
				if util.TrimWhitespace(cell) == "" {
					return m.renderer.recordDiagnostic(m.Node(), "Empty table header", "The header cell at index **"+strconv.Itoa(j)+"** cannot be empty.", true)
				}
			}
			firstRowCellCount = cellCount
		}
	}

	if len(rows) == 1 {
		return m.invalidDelimiter()
	}
	if len(rows) == 2 {
		return m.renderer.recordDiagnostic(m.Node(), "Empty table", "A table must have at least one row after the delimiter row.", true)
	}

	output := "<table><thead><tr>"
	for _, cell := range rows[0].Children() {
		output += `<th scope="col">` + m.cellContent(cell) + "</th>"
	}
	output += "</tr></thead><tbody>"
	for _, row := range rows[2:] {
		output += "<tr>"
		for _, cell := range row.Children() {
			output += "<td>" + m.cellContent(cell) + "</td>"
		}
		output += "</tr>"
	}
	output += "</tbody></table>"
	return output
}

func (m *markdownTable) invalidDelimiter() string {
	return m.renderer.recordDiagnostic(m.Node(), "Invalid table delimiter", "The table row at index **1** must be a delimiter row. Each cell must contain only **-**, at least three of them. Alignment is not supported.", true)
}

func (m *markdownTable) cellContent(cell ast.Node) string {
	return util.TrimWhitespace(m.renderer.renderChildren(m, cell.Children()))
}
