package component

import "github.com/umono-cms/compono/selector"

type static struct {
}

func (*static) Name() string {
	return "static"
}

func (*static) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (*static) DisallowParent() []string {
	return []string{"root"}
}
