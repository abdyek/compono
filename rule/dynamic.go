package rule

import "github.com/umono-cms/compono/selector"

type dynamic struct {
	name string
}

func NewDynamic(name string) Rule {
	return &dynamic{name: name}
}

func (d *dynamic) Name() string {
	return d.name
}

func (_ *dynamic) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (_ *dynamic) Rules() []Rule {
	return []Rule{}
}
