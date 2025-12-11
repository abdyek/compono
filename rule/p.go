package rule

import "github.com/umono-cms/compono/selector"

type p struct{}

func newP() Rule {
	return &p{}
}

func (_ *p) Name() string {
	return "p"
}

func (_ *p) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`^|\n\n`, `\n\n|\z`),
	}
}

func (_ *p) Rules() []Rule {
	return []Rule{
		newPContent(),
	}
}

type pContent struct{}

func newPContent() Rule {
	return &pContent{}
}

func (_ *pContent) Name() string {
	return "p-content"
}

func (_ *pContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *pContent) Rules() []Rule {
	return []Rule{
		newStrong(),
		newEm(),
		newInlineCompCall(),
		newParamRef(),
		newPlain(),
	}
}
