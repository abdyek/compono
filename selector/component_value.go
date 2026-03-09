package selector

import "bytes"

type componentAssignments struct {
	requireValue bool
}

func NewComponentAssignments(requireValue bool) Selector {
	return &componentAssignments{
		requireValue: requireValue,
	}
}

func (_ *componentAssignments) Name() string {
	return "component_assignments"
}

func (ca *componentAssignments) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	offset := 0

	for offset < len(source) {
		offset = skipComponentSpaces(source, offset)
		if offset >= len(source) {
			break
		}

		if source[offset] < 'a' || source[offset] > 'z' {
			offset++
			continue
		}

		start := offset
		nameEnd, ok := scanComponentParamName(source, offset)
		if !ok {
			offset++
			continue
		}
		offset = skipComponentSpaces(source, nameEnd)

		if offset >= len(source) || source[offset] != '=' {
			if ca.requireValue {
				offset = nameEnd
				continue
			}
			results = append(results, [2]int{start, nameEnd})
			offset = nameEnd
			continue
		}

		offset++
		offset = skipComponentSpaces(source, offset)

		valueEnd, ok := scanComponentValue(source, offset, ca.requireValue)
		if !ok {
			offset = nameEnd
			continue
		}

		results = append(results, [2]int{start, valueEnd})
		offset = valueEnd
	}

	return results
}

type arrayItems struct {
	allowParamRef bool
}

func NewArrayItems(allowParamRef bool) Selector {
	return &arrayItems{
		allowParamRef: allowParamRef,
	}
}

func (_ *arrayItems) Name() string {
	return "array_items"
}

func (ai *arrayItems) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	offset := 0

	for {
		offset = skipComponentSpaces(source, offset)
		if offset >= len(source) {
			break
		}

		start := offset
		valueEnd, ok := scanComponentValue(source, offset, ai.allowParamRef)
		if !ok {
			break
		}
		results = append(results, [2]int{start, valueEnd})

		offset = skipComponentSpaces(source, valueEnd)
		if offset >= len(source) {
			break
		}
		if source[offset] != ',' {
			break
		}
		offset++
	}

	return results
}

type arrayLiteral struct {
	allowParamRef bool
}

func NewArrayLiteral(allowParamRef bool) Selector {
	return &arrayLiteral{
		allowParamRef: allowParamRef,
	}
}

func (_ *arrayLiteral) Name() string {
	return "array_literal"
}

func (al *arrayLiteral) Select(source []byte, without ...[2]int) [][2]int {
	start := skipComponentSpaces(source, 0)
	if start >= len(source) || source[start] != '[' {
		return [][2]int{}
	}

	end, ok := scanArrayLiteral(source, start, al.allowParamRef)
	if !ok {
		return [][2]int{}
	}

	if skipComponentSpaces(source, end) != len(source) {
		return [][2]int{}
	}

	return [][2]int{{start, end}}
}

type paramRefIndexes struct{}

func NewParamRefIndexes() Selector {
	return &paramRefIndexes{}
}

func (_ *paramRefIndexes) Name() string {
	return "param_ref_indexes"
}

func (_ *paramRefIndexes) Select(source []byte, without ...[2]int) [][2]int {
	offset := bytes.Index(source, []byte("{{"))
	if offset == -1 {
		return [][2]int{}
	}

	offset += 2
	offset = skipComponentSpaces(source, offset)
	nameEnd, ok := scanComponentParamName(source, offset)
	if !ok {
		return [][2]int{}
	}

	start := nameEnd
	end := start
	for {
		next, ok := scanArrayIndex(source, end)
		if !ok {
			break
		}
		end = next
	}

	if end == start {
		return [][2]int{}
	}

	return [][2]int{{start, end}}
}

type arrayInner struct{}

func NewArrayInner() Selector {
	return &arrayInner{}
}

func (_ *arrayInner) Name() string {
	return "array_inner"
}

func (_ *arrayInner) Select(source []byte, without ...[2]int) [][2]int {
	start := skipComponentSpaces(source, 0)
	if start >= len(source) || source[start] != '[' {
		return [][2]int{}
	}

	end := len(source) - 1
	for end >= 0 && bytes.ContainsRune([]byte{' ', '\n', '\r', '\t'}, rune(source[end])) {
		end--
	}
	if end <= start || source[end] != ']' {
		return [][2]int{}
	}

	return [][2]int{{start + 1, end}}
}

func skipComponentSpaces(source []byte, offset int) int {
	for offset < len(source) {
		if !bytes.ContainsRune([]byte{' ', '\n', '\r', '\t'}, rune(source[offset])) {
			break
		}
		offset++
	}
	return offset
}

func scanComponentParamName(source []byte, offset int) (int, bool) {
	if offset >= len(source) || source[offset] < 'a' || source[offset] > 'z' {
		return 0, false
	}

	offset++
	for offset < len(source) {
		ch := source[offset]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			offset++
			continue
		}
		break
	}

	return offset, true
}

func scanComponentValue(source []byte, offset int, allowParamRef bool) (int, bool) {
	if offset >= len(source) {
		return 0, false
	}

	switch {
	case source[offset] == '"':
		return scanQuotedString(source, offset)
	case source[offset] == '[':
		return scanArrayLiteral(source, offset, allowParamRef)
	case source[offset] >= '0' && source[offset] <= '9':
		return scanNumberLiteral(source, offset)
	case hasComponentKeywordAt(source, offset, "true"):
		return offset + len("true"), true
	case hasComponentKeywordAt(source, offset, "false"):
		return offset + len("false"), true
	case source[offset] >= 'A' && source[offset] <= 'Z':
		return scanComponentName(source, offset)
	case allowParamRef && source[offset] >= 'a' && source[offset] <= 'z':
		return scanParamReferenceValue(source, offset)
	default:
		return 0, false
	}
}

func scanQuotedString(source []byte, offset int) (int, bool) {
	if source[offset] != '"' {
		return 0, false
	}

	offset++
	for offset < len(source) {
		if source[offset] == '"' {
			return offset + 1, true
		}
		offset++
	}

	return 0, false
}

func scanNumberLiteral(source []byte, offset int) (int, bool) {
	start := offset
	for offset < len(source) && source[offset] >= '0' && source[offset] <= '9' {
		offset++
	}

	if offset < len(source) && source[offset] == '.' {
		offset++
		for offset < len(source) && source[offset] >= '0' && source[offset] <= '9' {
			offset++
		}
	}

	return offset, offset > start
}

func scanComponentName(source []byte, offset int) (int, bool) {
	start := offset
	for offset < len(source) {
		ch := source[offset]
		if (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
			offset++
			continue
		}
		break
	}

	return offset, offset > start
}

func scanParamReferenceValue(source []byte, offset int) (int, bool) {
	end, ok := scanComponentParamName(source, offset)
	if !ok {
		return 0, false
	}

	for {
		next, ok := scanArrayIndex(source, end)
		if !ok {
			break
		}
		end = next
	}

	return end, true
}

func scanArrayLiteral(source []byte, offset int, allowParamRef bool) (int, bool) {
	if source[offset] != '[' {
		return 0, false
	}

	offset++
	for {
		offset = skipComponentSpaces(source, offset)
		if offset >= len(source) {
			return 0, false
		}
		if source[offset] == ']' {
			return offset + 1, true
		}

		valueEnd, ok := scanComponentValue(source, offset, allowParamRef)
		if !ok {
			return 0, false
		}
		offset = skipComponentSpaces(source, valueEnd)
		if offset >= len(source) {
			return 0, false
		}
		if source[offset] == ']' {
			return offset + 1, true
		}
		if source[offset] != ',' {
			return 0, false
		}
		offset++
	}
}

func scanArrayIndex(source []byte, offset int) (int, bool) {
	if offset >= len(source) || source[offset] != '[' {
		return 0, false
	}

	offset++
	start := offset
	for offset < len(source) && source[offset] >= '0' && source[offset] <= '9' {
		offset++
	}

	if offset == start || offset >= len(source) || source[offset] != ']' {
		return 0, false
	}

	return offset + 1, true
}

func hasComponentKeywordAt(source []byte, offset int, keyword string) bool {
	end := offset + len(keyword)
	if end > len(source) || string(source[offset:end]) != keyword {
		return false
	}
	if end >= len(source) {
		return true
	}
	next := source[end]
	return !((next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') || (next >= '0' && next <= '9') || next == '-' || next == '_' || next == '[')
}
