package component

import "github.com/umono-cms/compono/selector"

type h1 struct {
	scalable
}

func newH1() Component {
	return &h1{
		scalable: scalable{
			components: []Component{
				newH1Content(),
			},
		},
	}
}

func (h *h1) Name() string {
	return "h1"
}

func (h *h1) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEnd(`\s*# `, `\n|\z`),
	}
}

func (h *h1) Components() []Component {
	return h.components
}

type h1Content struct {
	scalable
}

func newH1Content() Component {
	return &h1Content{
		scalable: scalable{
			components: []Component{
				newEm(),
				newStrong(),
				newPlain(),
			},
		},
	}
}

func (*h1Content) Name() string {
	return "h1-content"
}

func (*h1Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\s*# `, `\n|\z`),
	}
}

func (h1c *h1Content) Components() []Component {
	return h1c.components
}
