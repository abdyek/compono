package rule

import (
	"github.com/umono-cms/compono/selector"
)

type list struct{}

func newList() Rule {
	return &list{}
}

func (_ *list) Name() string {
	return "list"
}

func (_ *list) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`(?m)(?:^[ \t]*-[ \t]+[^ \t\n][^\n]*(?:\n|\z))+`)
	return []selector.Selector{
		p,
	}
}

func (_ *list) Rules() []Rule {
	return []Rule{
		newListItem(),
	}
}

type listItem struct{}

func newListItem() Rule {
	return &listItem{}
}

func (_ *listItem) Name() string {
	return "list-item"
}

func (_ *listItem) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`(?m)^[^\n]+`)
	return []selector.Selector{
		p,
	}
}

func (_ *listItem) RemovedLines(source []byte) [][2]int {
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
		if i < lineEnd && source[i] == '-' {
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

func (_ *listItem) Rules() []Rule {
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
