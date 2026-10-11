package rule

import (
	"github.com/umono-cms/compono/selector"
)

type hr struct{}

func newHr() Rule {
	return &hr{}
}

func (_ *hr) Name() string {
	return "hr"
}

func (_ *hr) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`(?m)^[ \t]*---[ \t]*(?:\n|\z)`)
	return []selector.Selector{
		p,
	}
}

func (_ *hr) Rules() []Rule {
	return []Rule{}
}
