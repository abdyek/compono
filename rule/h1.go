package rule

import "github.com/umono-cms/compono/selector"

type h1 struct {
	scalable
}

func newH1() Rule {
	return &h1{
		scalable: scalable{
			rules: []Rule{
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

func (h *h1) Rules() []Rule {
	return h.rules
}

type h1Content struct {
	scalable
}

func newH1Content() Rule {
	return &h1Content{
		scalable: scalable{
			rules: []Rule{
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

func (h1c *h1Content) Rules() []Rule {
	return h1c.rules
}
