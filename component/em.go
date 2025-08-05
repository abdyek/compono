package component

import "github.com/umono-cms/compono/selector"

type em struct {
	scalable
}

func newEm() Component {
	return &em{
		scalable: scalable{
			components: []Component{
				newEmContent(),
			},
		},
	}
}

func (_ *em) Name() string {
	return "em"
}

func (_ *em) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEnd(`\*`, `\*`),
	}
}

func (e *em) Components() []Component {
	return e.components
}

type emContent struct {
	scalable
}

func newEmContent() Component {
	return &emContent{
		scalable: scalable{
			components: []Component{
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

func (ec *emContent) Components() []Component {
	return ec.components
}
