package rule

import "github.com/umono-cms/compono/selector"

type softBreak struct{}

func newSoftBreak() Rule {
	return &softBreak{}
}

func (_ *softBreak) Name() string {
	return "soft-break"
}

func (_ *softBreak) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\n`)
	return []selector.Selector{
		p,
	}
}

func (_ *softBreak) Rules() []Rule {
	return []Rule{}
}
