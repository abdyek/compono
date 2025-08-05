package component

import "github.com/umono-cms/compono/selector"

type h2 struct {
	scalable
}

func newH2() Component {
	return &h2{
		scalable: scalable{
			components: []Component{
				newH2Content(),
			},
		},
	}
}

func (h *h2) Name() string {
	return "h2"
}

func (h *h2) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEnd(`\s*## `, `\n|\z`),
	}
}

func (h *h2) Components() []Component {
	return h.components
}

type h2Content struct {
	scalable
}

func newH2Content() Component {
	return &h2Content{
		scalable: scalable{
			components: []Component{
				newStrong(),
				newEm(),
				newPlain(),
			},
		},
	}
}

func (h2c *h2Content) Name() string {
	return "h2-content"
}

func (h2c *h2Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\s*## `, `\n|\z`),
	}
}

func (h2c *h2Content) Components() []Component {
	return h2c.components
}
