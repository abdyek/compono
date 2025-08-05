package component

import "github.com/umono-cms/compono/selector"

type Root struct {
	scalable
}

func (*Root) Name() string {
	return "root"
}

func newRoot() Component {
	return &Root{}
}

func (*Root) Selectors() []selector.Selector {
	return []selector.Selector{}
}

func (*Root) Components() []Component {
	return []Component{
		newRootContent(),
	}
}

type rootContent struct {
	scalable
}

func newRootContent() Component {
	return &rootContent{
		scalable: scalable{
			components: []Component{
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
		selector.NewAll(),
	}
}

func (rc *rootContent) Components() []Component {
	return rc.components
}
