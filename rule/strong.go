package rule

import "github.com/umono-cms/compono/selector"

type strong struct{}

func newStrong() Rule {
	return &strong{}
}

func (_ *strong) Name() string {
	return "strong"
}

func (_ *strong) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			var result [][2]int
			for _, ind := range index {
				spans := emphasisSpans(source[ind[0]:ind[1]], 2)
				for _, span := range spans {
					result = append(result, [2]int{ind[0] + span[0], ind[0] + span[1]})
				}
			}
			return result
		}),
	}
}

func (_ *strong) Rules() []Rule {
	return []Rule{
		newStrongContent(),
	}
}

type strongContent struct{}

func newStrongContent() Rule {
	return &strongContent{}
}

func (sc *strongContent) Name() string {
	return "strong-content"
}

func (sc *strongContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			var result [][2]int
			for _, ind := range index {
				result = append(result, [2]int{ind[0] + 2, ind[1] - 2})
			}
			return result
		}),
	}
}

func (_ *strongContent) Rules() []Rule {
	return []Rule{
		newInlineCompCall(),
		newContextRef(),
		newParamRef(),
		newPlain(),
	}
}
