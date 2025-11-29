package rule

import "github.com/umono-cms/compono/selector"

type h1 struct{}

func newH1() Rule {
	return &h1{}
}

func (_ *h1) Name() string {
	return "h1"
}

func (_ *h1) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`(?m)[ \t]*# `, `\n|\z`)
	return []selector.Selector{
		seSelector,
	}
}

func (_ *h1) Rules() []Rule {
	return []Rule{
		newH1Content(),
	}
}

type h1Content struct{}

func newH1Content() Rule {
	return &h1Content{}
}

func (_ *h1Content) Name() string {
	return "h1-content"
}

func (_ *h1Content) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`(?m)[ \t]*# `, `\n|\z`),
	}
}

func (_ *h1Content) Rules() []Rule {
	return []Rule{
		newEm(),
		newStrong(),
		newInlineCompCall(),
		newParamRef(),
		newPlain(),
	}
}
