package selector

import "bytes"

type componentAssignments struct {
	requireValue bool
}

type mustacheCall struct {
	paramRefName bool
}

type contextMustacheCall struct{}

func NewMustacheCall(paramRefName bool) Selector {
	return &mustacheCall{
		paramRefName: paramRefName,
	}
}

func NewContextMustacheCall() Selector {
	return &contextMustacheCall{}
}

func (_ *mustacheCall) Name() string {
	return "mustache_call"
}

func (_ *contextMustacheCall) Name() string {
	return "context_mustache_call"
}

func (mc *mustacheCall) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	noSelected := filterNoSelected(without, len(source))

	for _, ns := range noSelected {
		offset := ns[0]
		for offset < ns[1] {
			start := bytes.Index(source[offset:ns[1]], []byte("{{"))
			if start == -1 {
				break
			}
			start += offset

			end, ok := scanMustacheCall(source, start, mc.paramRefName)
			if ok && end <= ns[1] {
				results = append(results, [2]int{start, end})
				offset = end
				continue
			}

			offset = start + 2
		}
	}

	return results
}

func (_ *contextMustacheCall) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	noSelected := filterNoSelected(without, len(source))

	for _, ns := range noSelected {
		offset := ns[0]
		for offset < ns[1] {
			start := bytes.Index(source[offset:ns[1]], []byte("{{"))
			if start == -1 {
				break
			}
			start += offset

			end, ok := scanContextMustacheCall(source, start)
			if ok && end <= ns[1] {
				results = append(results, [2]int{start, end})
				offset = end
				continue
			}

			offset = start + 2
		}
	}

	return results
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

type recordItems struct {
	allowParamRef bool
}

func NewRecordItems(allowParamRef bool) Selector {
	return &recordItems{
		allowParamRef: allowParamRef,
	}
}

func (_ *recordItems) Name() string {
	return "record_items"
}

func (ri *recordItems) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}
	offset := 0

	for {
		offset = skipComponentSpaces(source, offset)
		if offset >= len(source) {
			break
		}

		start := offset
		valueEnd, ok := scanRecordEntry(source, offset, ri.allowParamRef)
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

type recordLiteral struct {
	allowParamRef bool
}

func NewRecordLiteral(allowParamRef bool) Selector {
	return &recordLiteral{
		allowParamRef: allowParamRef,
	}
}

func (_ *recordLiteral) Name() string {
	return "record_literal"
}

func (rl *recordLiteral) Select(source []byte, without ...[2]int) [][2]int {
	start := skipComponentSpaces(source, 0)
	if start >= len(source) || source[start] != '{' {
		return [][2]int{}
	}

	end, ok := scanRecordLiteral(source, start, rl.allowParamRef)
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

type recordInner struct{}

func NewRecordInner() Selector {
	return &recordInner{}
}

func (_ *recordInner) Name() string {
	return "record_inner"
}

func (_ *recordInner) Select(source []byte, without ...[2]int) [][2]int {
	start := skipComponentSpaces(source, 0)
	if start >= len(source) || source[start] != '{' {
		return [][2]int{}
	}

	end := len(source) - 1
	for end >= 0 && bytes.ContainsRune([]byte{' ', '\n', '\r', '\t'}, rune(source[end])) {
		end--
	}
	if end <= start || source[end] != '}' {
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
	case hasContextKeywordAt(source, offset):
		return scanContextValue(source, offset)
	case source[offset] == '"':
		return scanQuotedString(source, offset)
	case source[offset] == '[':
		return scanArrayLiteral(source, offset, allowParamRef)
	case source[offset] == '{':
		return scanRecordLiteral(source, offset, allowParamRef)
	case source[offset] == '-' || (source[offset] >= '0' && source[offset] <= '9'):
		return scanIntegerLiteral(source, offset)
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

func hasContextKeywordAt(source []byte, offset int) bool {
	return offset+7 <= len(source) && string(source[offset:offset+7]) == "context"
}

func scanMustacheCall(source []byte, offset int, paramRefName bool) (int, bool) {
	if offset+1 >= len(source) || source[offset] != '{' || source[offset+1] != '{' {
		return 0, false
	}

	offset += 2
	offset = skipComponentSpaces(source, offset)

	var end int
	var ok bool
	if paramRefName {
		end, ok = scanParamReferenceValue(source, offset)
	} else {
		end, ok = scanComponentName(source, offset)
	}
	if !ok {
		return 0, false
	}
	offset = end

	for {
		offset = skipComponentSpaces(source, offset)
		if offset >= len(source) {
			return 0, false
		}

		if offset+1 < len(source) && source[offset] == '}' && source[offset+1] == '}' {
			return offset + 2, true
		}

		nameEnd, ok := scanComponentParamName(source, offset)
		if !ok {
			return 0, false
		}
		offset = skipComponentSpaces(source, nameEnd)
		if offset >= len(source) || source[offset] != '=' {
			return 0, false
		}

		offset++
		offset = skipComponentSpaces(source, offset)
		valueEnd, ok := scanComponentValue(source, offset, true)
		if !ok {
			return 0, false
		}
		offset = valueEnd
	}
}

func scanContextMustacheCall(source []byte, offset int) (int, bool) {
	if offset+1 >= len(source) || source[offset] != '{' || source[offset+1] != '{' {
		return 0, false
	}

	offset += 2
	offset = skipComponentSpaces(source, offset)
	end, ok := scanContextValue(source, offset)
	if !ok {
		return 0, false
	}

	offset = skipComponentSpaces(source, end)
	if offset+1 >= len(source) || source[offset] != '}' || source[offset+1] != '}' {
		return 0, false
	}

	return offset + 2, true
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

func scanIntegerLiteral(source []byte, offset int) (int, bool) {
	start := offset
	if source[offset] == '-' {
		offset++
		if offset >= len(source) || source[offset] < '0' || source[offset] > '9' {
			return 0, false
		}
	}

	for offset < len(source) && source[offset] >= '0' && source[offset] <= '9' {
		offset++
	}

	if offset < len(source) && source[offset] == '.' {
		return 0, false
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
		next := skipComponentSpaces(source, end)
		switch {
		case next < len(source) && source[next] == '.':
			next++
			next = skipComponentSpaces(source, next)
			keyEnd, ok := scanComponentParamName(source, next)
			if !ok {
				return 0, false
			}
			end = keyEnd
		default:
			indexEnd, ok := scanArrayIndex(source, next)
			if !ok {
				return end, true
			}
			end = indexEnd
		}
	}
}

func scanContextValue(source []byte, offset int) (int, bool) {
	if !hasContextKeywordAt(source, offset) {
		return 0, false
	}

	offset += len("context")
	if offset >= len(source) || source[offset] != '(' {
		return 0, false
	}

	offset++
	offset = skipComponentSpaces(source, offset)
	start := offset
	for offset < len(source) && source[offset] != ')' {
		ch := source[offset]
		if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' {
			offset++
			continue
		}
		if ch == '/' || ch == '-' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			offset++
			continue
		}
		return 0, false
	}
	if offset >= len(source) || !isValidContextKey(string(bytes.TrimSpace(source[start:offset]))) {
		return 0, false
	}

	offset++
	for {
		next := skipComponentSpaces(source, offset)
		switch {
		case next < len(source) && source[next] == '.':
			next++
			next = skipComponentSpaces(source, next)
			keyEnd, ok := scanComponentParamName(source, next)
			if !ok {
				return 0, false
			}
			offset = keyEnd
		default:
			indexEnd, ok := scanArrayIndex(source, next)
			if !ok {
				return offset, true
			}
			offset = indexEnd
		}
	}
}

func isValidContextKey(key string) bool {
	if len(key) == 0 || key[0] == '/' || key[len(key)-1] == '/' {
		return false
	}
	for i := 0; i < len(key); i++ {
		ch := key[i]
		if ch == '/' {
			if i+1 < len(key) && key[i+1] == '/' {
				return false
			}
			continue
		}
		if ch == '-' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			continue
		}
		return false
	}
	return true
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

func scanRecordLiteral(source []byte, offset int, allowParamRef bool) (int, bool) {
	if source[offset] != '{' {
		return 0, false
	}

	offset++
	for {
		offset = skipComponentSpaces(source, offset)
		if offset >= len(source) {
			return 0, false
		}
		if source[offset] == '}' {
			return offset + 1, true
		}

		entryEnd, ok := scanRecordEntry(source, offset, allowParamRef)
		if !ok {
			return 0, false
		}

		offset = skipComponentSpaces(source, entryEnd)
		if offset >= len(source) {
			return 0, false
		}
		if source[offset] == '}' {
			return offset + 1, true
		}
		if source[offset] != ',' {
			return 0, false
		}
		offset++
	}
}

func scanRecordEntry(source []byte, offset int, allowParamRef bool) (int, bool) {
	keyEnd, ok := scanComponentParamName(source, offset)
	if !ok {
		return 0, false
	}

	offset = skipComponentSpaces(source, keyEnd)
	if offset >= len(source) || source[offset] != ':' {
		return 0, false
	}

	offset++
	offset = skipComponentSpaces(source, offset)
	return scanComponentValue(source, offset, allowParamRef)
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
