package rule

import "github.com/umono-cms/compono/selector"

type Root struct{}

func (_ *Root) Name() string {
	return "root"
}

func newRoot() Rule {
	return &Root{}
}

func (_ *Root) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (_ *Root) Rules() []Rule {
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
	return []selector.Selector{
		selector.NewUntilFirstMatch(`\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*\s*\n`),
	}
}

func (_ *rootContent) Rules() []Rule {
	return []Rule{
		newH2(),
		newH1(),
		newP(),
		newBlockCompCall(),
	}
}
