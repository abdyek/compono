package rule

import (
	"github.com/umono-cms/compono/selector"
)

type plain struct {
	scalable
}

func newPlain() Rule {
	return &plain{}
}

func (*plain) Name() string {
	return "plain"
}

func (*plain) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (*plain) Rules() []Rule {
	return []Rule{}
}
