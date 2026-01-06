package rule

import (
	"sort"

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
	se, _ := selector.NewStartEnd("```", "```")
	return []selector.Selector{
		selector.NewFilter(se, func(source []byte, index [][2]int) [][2]int {
			if len(index) == 0 {
				return [][2]int{}
			}

			noMirage := filterOutMirage(index)
			filtered := [][2]int{}

			for _, ind := range noMirage {
				start, end := ind[0], ind[1]

				leftOK := true
				for i := start - 1; i >= 0 && source[i] != '\n'; i-- {
					if source[i] != ' ' && source[i] != '\t' {
						leftOK = false
						break
					}
				}

				rightOK := true
				for i := end; i < len(source) && source[i] != '\n'; i++ {
					if source[i] != ' ' && source[i] != '\t' {
						rightOK = false
						break
					}
				}

				leftOfEndOK := true
				for i := end - 4; i >= 0 && source[i] != '\n'; i-- {
					if source[i] != ' ' && source[i] != '\t' {
						leftOfEndOK = false
						break
					}
				}

				rightOfStartOK := true
				mustWhitespace := false
				for i := start + 3; i < len(source) && source[i] != '\n'; i++ {
					if mustWhitespace && source[i] != ' ' && source[i] != '\t' {
						rightOfStartOK = false
						break
					}
					if source[i] == ' ' || source[i] == '\t' {
						mustWhitespace = true
					}
				}

				if leftOK && rightOK && leftOfEndOK && rightOfStartOK {
					filtered = append(filtered, ind)
				}
			}
			return filtered
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
		selector.NewStartEndInner("```", `\n`),
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
		selector.NewStartEndInner(`\n`, "```"),
	}
}

func (_ *codeBlockContent) Rules() []Rule {
	return []Rule{
		newPlain(),
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
	se, _ := selector.NewStartEnd("`", "`")
	return []selector.Selector{
		selector.NewFilter(se, func(source []byte, index [][2]int) [][2]int {
			if len(index) == 0 {
				return [][2]int{}
			}

			noMirage := filterOutMirage(index)
			filtered := [][2]int{}

			for _, ind := range noMirage {
				start, end := ind[0], ind[1]

				noNewline := true
				for i := start; i < end; i++ {
					if source[i] == '\n' {
						noNewline = true
						break
					}
				}

				single := true
				if (start > 0 && source[start-1] == '`') || (end < len(source) && source[end] == '`') {
					single = false
				}

				if noNewline && single {
					filtered = append(filtered, ind)
				}
			}

			return filtered
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
		selector.NewStartEndInner("`", "`"),
	}
}

func (_ *inlineCodeContent) Rules() []Rule {
	return []Rule{
		newRaw(),
	}
}

func filterOutMirage(indexes [][2]int) [][2]int {
	if len(indexes) == 0 {
		return indexes
	}

	sorted := make([][2]int, len(indexes))
	copy(sorted, indexes)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i][0] < sorted[j][0]
	})

	valid := make([][2]int, 0)

	for _, interval := range sorted {
		if len(valid) > 0 && interval[0] < valid[len(valid)-1][1] {
			continue
		}
		valid = append(valid, interval)
	}

	return valid
}
