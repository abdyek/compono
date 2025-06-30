package component

import "github.com/umono-cms/compono/selector"

type h1 struct {
}

func (h *h1) Name() string {
	return "h1"
}

func (h *h1) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEnd(`\s*# `, `\n|\z`),
	}
}

func (h *h1) DisallowParent() []string {
	return []string{"p", "static"}
}
