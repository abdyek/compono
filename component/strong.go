package component

import "github.com/umono-cms/compono/selector"

type strong struct {
	scalable
}

func newStrong() Component {
	return &strong{
		scalable: scalable{
			components: []Component{
				newStrongContent(),
			},
		},
	}
}

func (s *strong) Name() string {
	return "strong"
}

func (s *strong) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEnd(`\*\*`, `\*\*`),
	}
}

func (s *strong) Components() []Component {
	return s.components
}

type strongContent struct {
	scalable
}

func newStrongContent() Component {
	return &strongContent{
		scalable: scalable{
			components: []Component{
				newPlain(),
			},
		},
	}
}

func (sc *strongContent) Name() string {
	return "strong-content"
}

func (sc *strongContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\*\*`, `\*\*`),
	}
}

func (sc *strongContent) Components() []Component {
	return sc.components
}
