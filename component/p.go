package component

import "github.com/umono-cms/compono/selector"

type p struct {
	scalable
}

func newP() Component {
	return &p{
		scalable: scalable{
			components: []Component{
				newPContent(),
			},
		},
	}
}

func (*p) Name() string {
	return "p"
}

func (*p) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (p *p) Components() []Component {
	return p.components
}

type pContent struct {
	scalable
}

func newPContent() Component {
	return &pContent{
		scalable: scalable{
			components: []Component{
				newStrong(),
				newEm(),
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

func (pc *pContent) Components() []Component {
	return pc.components
}
