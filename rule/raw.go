package rule

import (
	"github.com/umono-cms/compono/selector"
)

type raw struct{}

func newRaw() Rule {
	return &raw{}
}

func (_ *raw) Name() string {
	return "raw"
}

func (_ *raw) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *raw) Rules() []Rule {
	return []Rule{}
}
