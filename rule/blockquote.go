package rule

import (
	"github.com/umono-cms/compono/selector"
)

type blockquote struct{}

func newBlockquote() Rule {
	return &blockquote{}
}

func (_ *blockquote) Name() string {
	return "blockquote"
}

func (_ *blockquote) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`(?m)(?:^[ \t]*>(?:[ \t][^\n]*)?(?:\n|\z))+`)
	return []selector.Selector{
		selector.NewFilter(p, func(source []byte, index [][2]int) [][2]int {
			filtered := [][2]int{}

			for _, ind := range index {
				start, end := ind[0], ind[1]
				content := source[start:end]

				hasQuoteLine := false

				pos := 0
				for pos < len(content) {
					lineEnd := pos
					for lineEnd < len(content) && content[lineEnd] != '\n' {
						lineEnd++
					}
					if lineEnd < len(content) {
						lineEnd++
					}

					i := pos
					for i < lineEnd && (content[i] == ' ' || content[i] == '\t') {
						i++
					}
					if i < lineEnd && content[i] == '>' {
						i++
						for i < lineEnd && (content[i] == ' ' || content[i] == '\t') {
							i++
						}
						for j := i; j < lineEnd; j++ {
							c := content[j]
							if c != ' ' && c != '\t' && c != '\n' {
								hasQuoteLine = true
								break
							}
						}
					}

					pos = lineEnd
				}

				if hasQuoteLine {
					filtered = append(filtered, ind)
				}
			}

			return filtered
		}),
	}
}

func (_ *blockquote) RemovedLines(source []byte) [][2]int {
	removed := [][2]int{}

	pos := 0
	for pos < len(source) {
		lineStart := pos
		lineEnd := pos
		for lineEnd < len(source) && source[lineEnd] != '\n' {
			lineEnd++
		}
		if lineEnd < len(source) {
			lineEnd++
		}

		i := lineStart
		for i < lineEnd && (source[i] == ' ' || source[i] == '\t') {
			i++
		}
		if i < lineEnd && source[i] == '>' {
			i++
		}
		for i < lineEnd && (source[i] == ' ' || source[i] == '\t') {
			i++
		}

		removed = append(removed, [2]int{lineStart, i})
		pos = lineEnd
	}

	return removed
}

func (_ *blockquote) Rules() []Rule {
	return []Rule{
		newP(),
	}
}
