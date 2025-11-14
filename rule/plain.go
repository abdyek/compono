package rule

import (
	"github.com/umono-cms/compono/selector"
)

type plain struct{}

func newPlain() Rule {
	return &plain{}
}

func (_ *plain) Name() string {
	return "plain"
}

func (_ *plain) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *plain) Rules() []Rule {
	return []Rule{}
}
