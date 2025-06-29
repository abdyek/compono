package component

import "github.com/umono-cms/compono/selector"

type p struct {
}

func (*p) Name() string {
	return "p"
}

func (*p) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (*p) DisallowParent() []string {
	return []string{}
}
