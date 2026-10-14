package rule

import "github.com/umono-cms/compono/selector"

type em struct{}

func newEm() Rule {
	return &em{}
}

func (_ *em) Name() string {
	return "em"
}

func (_ *em) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			var result [][2]int
			for _, ind := range index {
				spans := emphasisSpans(source[ind[0]:ind[1]], 1)
				for _, span := range spans {
					result = append(result, [2]int{ind[0] + span[0], ind[0] + span[1]})
				}
			}
			return result
		}),
	}
}

func (_ *em) Rules() []Rule {
	return []Rule{
		newEmContent(),
	}
}

type emContent struct{}

func newEmContent() Rule {
	return &emContent{}
}

func (_ *emContent) Name() string {
	return "em-content"
}

func (_ *emContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {
			var result [][2]int
			for _, ind := range index {
				result = append(result, [2]int{ind[0] + 1, ind[1] - 1})
			}
			return result
		}),
	}
}

func (_ *emContent) Rules() []Rule {
	return []Rule{
		newInlineCompCall(),
		newContextRef(),
		newParamRef(),
		newInvalidExpression(),
		newUnclosedExpression(),
		newPlain(),
	}
}
