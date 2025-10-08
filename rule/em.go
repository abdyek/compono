package rule

import "github.com/umono-cms/compono/selector"

type em struct {
	scalable
}

func newEm() Rule {
	return &em{
		scalable: scalable{
			rules: []Rule{
				newEmContent(),
			},
		},
	}
}

func (_ *em) Name() string {
	return "em"
}

func (_ *em) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\*`, `\*`)
	return []selector.Selector{
		seSelector,
	}
}

func (e *em) Rules() []Rule {
	return e.rules
}

type emContent struct {
	scalable
}

func newEmContent() Rule {
	return &emContent{
		scalable: scalable{
			rules: []Rule{
				newPlain(),
			},
		},
	}
}

func (_ *emContent) Name() string {
	return "em-content"
}

func (_ *emContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\*`, `\*`),
	}
}

func (ec *emContent) Rules() []Rule {
	return ec.rules
}
