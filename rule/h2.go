package rule

import "github.com/umono-cms/compono/selector"

type h2 struct {
	scalable
}

func newH2() Rule {
	return &h2{
		scalable: scalable{
			rules: []Rule{
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

func (h *h2) Rules() []Rule {
	return h.rules
}

type h2Content struct {
	scalable
}

func newH2Content() Rule {
	return &h2Content{
		scalable: scalable{
			rules: []Rule{
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

func (h2c *h2Content) Rules() []Rule {
	return h2c.rules
}
