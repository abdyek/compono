package selector

import (
	"fmt"
	"regexp"
)

type pattern struct {
	regex *regexp.Regexp
}

func NewPattern(ptrn string) (Selector, error) {
	re, err := regexp.Compile(ptrn)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	return &pattern{
		regex: re,
	}, nil
}

func (_ *pattern) Name() string {
	return "pattern"
}

func (p *pattern) Select(source []byte, without ...[2]int) [][2]int {
	if len(source) == 0 {
		return [][2]int{}
	}

	results := [][2]int{}
	noSelected := filterNoSelected(without, len(source))
	for _, ns := range noSelected {
		piece := string(source[ns[0]:ns[1]])
		for _, i := range p.regex.FindAllStringIndex(piece, -1) {
			results = append(results, [2]int{ns[0] + i[0], ns[0] + i[1]})
		}
	}
	return results
}
