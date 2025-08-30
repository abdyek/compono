package rule

import "github.com/umono-cms/compono/selector"

type Root struct {
	scalable
}

func (*Root) Name() string {
	return "root"
}

func newRoot() Rule {
	return &Root{}
}

func (*Root) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (*Root) Rules() []Rule {
	return []Rule{
		newRootContent(),
	}
}

type rootContent struct {
	scalable
}

func newRootContent() Rule {
	return &rootContent{
		scalable: scalable{
			rules: []Rule{
				newH2(),
				newH1(),
				newP(),
			},
		},
	}
}

func (*rootContent) Name() string {
	return "root-content"
}

func (*rootContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewUntilFirstMatch(`\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*\s*\n`),
	}
}

func (rc *rootContent) Rules() []Rule {
	return rc.rules
}
