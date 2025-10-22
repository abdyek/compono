package rule

import "github.com/umono-cms/compono/selector"

type p struct {
	scalable
}

func newP() Rule {
	return &p{
		scalable: scalable{
			rules: []Rule{
				newPContent(),
			},
		},
	}
}

func (*p) Name() string {
	return "p"
}

func (*p) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`^`, `\n\n|\z`)
	return []selector.Selector{
		seSelector,
	}
}

func (p *p) Rules() []Rule {
	return p.rules
}

type pContent struct {
	scalable
}

func newPContent() Rule {
	return &pContent{
		scalable: scalable{
			rules: []Rule{
				newStrong(),
				newEm(),
				newInlineCompCall(),
				newPlain(),
			},
		},
	}
}

func (pc *pContent) Name() string {
	return "p-content"
}

func (pc *pContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (pc *pContent) Rules() []Rule {
	return pc.rules
}
