package rule

import "github.com/umono-cms/compono/selector"

type p struct{}

func newP() Rule {
	return &p{}
}

func (_ *p) Name() string {
	return "p"
}

func (p *p) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewFilter(selector.NewAll(), func(source []byte, index [][2]int) [][2]int {

			if len(index) == 0 {
				return [][2]int{}
			}

			res := [][2]int{}

			for _, ind := range index {
				start, end := ind[0], ind[1]
				content := source[start:end]

				current := 0
				for current < len(content) {
					sepIdx := -1
					for i := current; i < len(content)-1; i++ {
						if content[i] == '\n' && content[i+1] == '\n' {
							sepIdx = i
							break
						}
					}

					if sepIdx == -1 {
						if current < len(content) {
							segment := content[current:]
							if !p.isEmpty(segment) {
								res = append(res, [2]int{start + current, start + len(content)})
							}
						}
						break
					}

					if current < sepIdx {
						segment := content[current:sepIdx]
						if !p.isEmpty(segment) {
							res = append(res, [2]int{start + current, start + sepIdx})
						}
					}

					current = sepIdx + 2
				}
			}

			return res
		}),
	}
}

func (_ *p) isEmpty(segment []byte) bool {
	for _, b := range segment {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			return false
		}
	}
	return true
}

func (_ *p) Rules() []Rule {
	return []Rule{
		newPContent(),
	}
}

type pContent struct{}

func newPContent() Rule {
	return &pContent{}
}

func (_ *pContent) Name() string {
	return "p-content"
}

func (_ *pContent) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *pContent) Rules() []Rule {
	return []Rule{
		newLink(),
		newStrong(),
		newEm(),
		newInlineCode(),
		newInlineCompCall(),
		newParamRef(),
		newSoftBreak(),
		newPlain(),
	}
}
