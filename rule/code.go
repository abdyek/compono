package rule

import (
	"bytes"

	"github.com/umono-cms/compono/selector"
)

type codeBlock struct{}

func newCodeBlock() Rule {
	return &codeBlock{}
}

func (_ *codeBlock) Name() string {
	return "code-block"
}

func (cb *codeBlock) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			result := [][2]int{}
			for _, ind := range index {
				s, e := ind[0], ind[1]
				nextLinePos := func(pos int) int {
					if pos < e && source[pos] == '\n' {
						return pos + 1
					}
					return pos
				}
				line := s
				for line < e {
					lineEnd := line
					for lineEnd < e && source[lineEnd] != '\n' {
						lineEnd++
					}
					trimmed := bytes.Trim(source[line:lineEnd], " \t")
					n := 0
					for n < len(trimmed) && trimmed[n] == '`' {
						n++
					}
					if n < 3 {
						line = nextLinePos(lineEnd)
						continue
					}
					validOpening := true
					for _, c := range trimmed[n:] {
						if c == ' ' || c == '\t' {
							validOpening = false
							break
						}
					}
					if !validOpening {
						line = nextLinePos(lineEnd)
						continue
					}
					openStart := line
					for openStart < lineEnd && source[openStart] != '`' {
						openStart++
					}
					next := nextLinePos(lineEnd)
					found := false
					for next < e {
						closeLineEnd := next
						for closeLineEnd < e && source[closeLineEnd] != '\n' {
							closeLineEnd++
						}
						closeTrimmed := bytes.Trim(source[next:closeLineEnd], " \t")
						if len(closeTrimmed) == n {
							allBT := true
							for _, c := range closeTrimmed {
								if c != '`' {
									allBT = false
									break
								}
							}
							if allBT {
								closeFenceStart := next
								for closeFenceStart < closeLineEnd && source[closeFenceStart] != '`' {
									closeFenceStart++
								}
								endPos := closeFenceStart + n
								result = append(result, [2]int{openStart, endPos})
								line = nextLinePos(closeLineEnd)
								found = true
								break
							}
						}
						next = nextLinePos(closeLineEnd)
					}
					if !found {
						line = nextLinePos(lineEnd)
					}
				}
			}
			return result
		}),
	}
}

func (_ *codeBlock) Rules() []Rule {
	return []Rule{
		newCodeBlockLang(),
		newCodeBlockContent(),
	}
}

type codeBlockLang struct{}

func newCodeBlockLang() Rule {
	return &codeBlockLang{}
}

func (_ *codeBlockLang) Name() string {
	return "code-block-lang"
}

func (_ *codeBlockLang) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			result := [][2]int{}
			for _, ind := range index {
				s, e := ind[0], ind[1]
				n := 0
				for s+n < e && source[s+n] == '`' {
					n++
				}
				lineEnd := s
				for lineEnd < e && source[lineEnd] != '\n' {
					lineEnd++
				}
				langEnd := lineEnd
				for langEnd > s+n && (source[langEnd-1] == ' ' || source[langEnd-1] == '\t') {
					langEnd--
				}
				if langEnd > s+n {
					result = append(result, [2]int{s + n, langEnd})
				}
			}
			return result
		}),
	}
}

func (_ *codeBlockLang) Rules() []Rule {
	return []Rule{}
}

type codeBlockContent struct{}

func newCodeBlockContent() Rule {
	return &codeBlockContent{}
}

func (_ *codeBlockContent) Name() string {
	return "code-block-content"
}

func (_ *codeBlockContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			result := [][2]int{}
			for _, ind := range index {
				s, e := ind[0], ind[1]
				firstNL := -1
				lastNL := -1
				for i := s; i < e; i++ {
					if source[i] == '\n' {
						if firstNL == -1 {
							firstNL = i
						}
						lastNL = i
					}
				}
				if firstNL != -1 && lastNL > firstNL {
					result = append(result, [2]int{firstNL + 1, lastNL + 1})
				}
			}
			return result
		}),
	}
}

func (_ *codeBlockContent) Rules() []Rule {
	return []Rule{
		newRaw(),
	}
}

type inlineCode struct{}

func newInlineCode() Rule {
	return &inlineCode{}
}

func (_ *inlineCode) Name() string {
	return "inline-code"
}

func (_ *inlineCode) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			result := [][2]int{}
			for _, ind := range index {
				s, e := ind[0], ind[1]
				codes, _ := selector.InlineUnits(source[s:e])
				for _, c := range codes {
					result = append(result, [2]int{s + c[0], s + c[1]})
				}
			}
			return result
		}),
	}
}

func (_ *inlineCode) Rules() []Rule {
	return []Rule{
		newInlineCodeContent(),
	}
}

type inlineCodeContent struct{}

func newInlineCodeContent() Rule {
	return &inlineCodeContent{}
}

func (_ *inlineCodeContent) Name() string {
	return "inline-code-content"
}

func (_ *inlineCodeContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			result := [][2]int{}
			for _, ind := range index {
				s, e := ind[0], ind[1]
				result = append(result, [2]int{s + 1, e - 1})
			}
			return result
		}),
	}
}

func (_ *inlineCodeContent) Rules() []Rule {
	return []Rule{
		newRaw(),
	}
}
