package rule

import "github.com/umono-cms/compono/selector"

type root struct{}

func (_ *root) Name() string {
	return "root"
}

func NewRoot() Rule {
	return &root{}
}

func (_ *root) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (_ *root) Rules() []Rule {
	return []Rule{
		newRootContent(),
		newLocalCompDefWrapper(),
	}
}

type rootContent struct{}

func newRootContent() Rule {
	return &rootContent{}
}

func (_ *rootContent) Name() string {
	return "root-content"
}

func (_ *rootContent) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`^`, `\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*|\z`)
	return []selector.Selector{
		seli,
	}
}

func (_ *rootContent) Rules() []Rule {
	return []Rule{
		newH2(),
		newH1(),
		newBlockCompCall(),
		newP(),
	}
}
