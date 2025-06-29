package component

import "github.com/umono-cms/compono/selector"

type Root struct {
}

func (*Root) Name() string {
	return "root"
}

func (*Root) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (*Root) DisallowParent() []string {
	return []string{}
}
