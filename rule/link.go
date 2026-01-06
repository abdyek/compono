package rule

import "github.com/umono-cms/compono/selector"

type link struct{}

func newLink() Rule {
	return &link{}
}

func (_ *link) Name() string {
	return "link"
}

func (_ *link) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\[`, `\)`)
	return []selector.Selector{
		selector.NewFilter(seSelector, func(source []byte, index [][2]int) [][2]int {
			filtered := [][2]int{}
			for _, ind := range index {
				content := source[ind[0]:ind[1]]
				hasClosingBracket := false
				parenStart := -1
				for i := 0; i < len(content); i++ {
					if content[i] == ']' && i+1 < len(content) && content[i+1] == '(' {
						hasClosingBracket = true
						parenStart = i + 1
						break
					}
				}
				if hasClosingBracket && parenStart > 0 {
					filtered = append(filtered, ind)
				}
			}
			return filtered
		}),
	}
}

func (_ *link) Rules() []Rule {
	return []Rule{
		newLinkText(),
		newLinkURL(),
	}
}

type linkText struct{}

func newLinkText() Rule {
	return &linkText{}
}

func (_ *linkText) Name() string {
	return "link-text"
}

func (_ *linkText) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\[`, `\]`),
	}
}

func (_ *linkText) Rules() []Rule {
	return []Rule{
		newStrong(),
		newEm(),
		newInlineCode(),
		newPlain(),
	}
}

type linkURL struct{}

func newLinkURL() Rule {
	return &linkURL{}
}

func (_ *linkURL) Name() string {
	return "link-url"
}

func (_ *linkURL) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\]\(`, `\)`),
	}
}

func (_ *linkURL) Rules() []Rule {
	return []Rule{}
}
