package rule

import "github.com/umono-cms/compono/selector"

// LineRemover is implemented by rules whose source has lines that are removed
// before it is parsed. The parser parses the source without the returned
// [start, end) ranges, and node ranges still refer to the original source.
type LineRemover interface {
	RemovedLines(source []byte) [][2]int
}

// commentLines returns the [start, end) ranges of the comment lines of
// source, each with its line break.
func commentLines(source []byte) [][2]int {
	codeSpans := newCodeBlock().Selectors()[0].Select(source)

	comments := [][2]int{}
	protectedUntil := 0

	for start := 0; start < len(source); {
		end := start
		for end < len(source) && source[end] != '\n' {
			end++
		}
		if end < len(source) {
			end++
		}

		inCode := false
		for _, span := range codeSpans {
			if start >= span[0] && start < span[1] {
				inCode = true
				break
			}
		}

		isComment := false
		if start >= protectedUntil && !inCode {
			i := start
			for i < end && (source[i] == ' ' || source[i] == '\t') {
				i++
			}
			if i+1 < end && source[i] == '/' && source[i+1] == '/' {
				isComment = true
				comments = append(comments, [2]int{start, end})
			}
		}

		if !isComment && !inCode {
			i := start
			if protectedUntil > i {
				i = protectedUntil
			}
			for i+1 < end {
				if source[i] == '{' && source[i+1] == '{' {
					s := i
					unitEnd, ok := selector.MustacheCallEnd(source, s)
					if !ok {
						unitEnd = s + 2
						for j := s + 2; j+1 < len(source); j++ {
							if source[j] == '{' && source[j+1] == '{' {
								break
							}
							if source[j] == '}' && source[j+1] == '}' {
								unitEnd = j + 2
								break
							}
						}
					}
					if unitEnd > protectedUntil {
						protectedUntil = unitEnd
					}
					if unitEnd > i {
						i = unitEnd
					} else {
						i++
					}
					continue
				}
				i++
			}
		}

		start = end
	}

	return comments
}
