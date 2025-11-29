package rule

import "github.com/umono-cms/compono/selector"

type h2 struct{}

func newH2() Rule {
	return &h2{}
}

func (_ *h2) Name() string {
	return "h2"
}

func (_ *h2) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`(?m)[ \t]*## `, `\n|\z`)
	return []selector.Selector{
		seSelector,
	}
}

func (_ *h2) Rules() []Rule {
	return []Rule{
		newH2Content(),
	}
}

type h2Content struct{}

func newH2Content() Rule {
	return &h2Content{}
}

func (_ *h2Content) Name() string {
	return "h2-content"
}

func (_ *h2Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`(?m)[ \t]*## `, `\n|\z`),
	}
}

func (_ *h2Content) Rules() []Rule {
	return []Rule{
		newStrong(),
		newEm(),
		newInlineCompCall(),
		newParamRef(),
		newPlain(),
	}
}
