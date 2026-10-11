package selector

type quotedStringInner struct{}

func NewQuotedStringInner() *quotedStringInner {
	return &quotedStringInner{}
}

func (_ *quotedStringInner) Name() string {
	return "quoted_string_inner"
}

func (_ *quotedStringInner) Select(source []byte, without ...[2]int) [][2]int {
	if len(source) == 0 {
		return nil
	}

	var results [][2]int

	i := 0
OUTER:
	for i < len(source) {
		for i < len(source) && isComponentSpace(source[i]) {
			i++
		}
		if i >= len(source) || source[i] != '"' {
			break
		}

		quoteStart := i
		endAbs, ok := scanQuotedString(source, quoteStart)
		if !ok {
			break
		}

		contentStart := quoteStart + 1
		contentEnd := endAbs - 1

		for _, w := range without {
			if contentStart < w[1] && contentEnd > w[0] {
				i = endAbs
				continue OUTER
			}
		}

		results = append(results, [2]int{contentStart, contentEnd})
		i = endAbs
	}

	return results
}
