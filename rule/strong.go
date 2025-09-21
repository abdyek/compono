package rule

import "github.com/umono-cms/compono/selector"

type strong struct {
	scalable
}

func newStrong() Rule {
	return &strong{
		scalable: scalable{
			rules: []Rule{
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

func (s *strong) Rules() []Rule {
	return s.rules
}

type strongContent struct {
	scalable
}

func newStrongContent() Rule {
	return &strongContent{
		scalable: scalable{
			rules: []Rule{
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

func (sc *strongContent) Rules() []Rule {
	return sc.rules
}
