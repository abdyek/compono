package selector

import (
	"regexp"
)

type startEndLeftInner struct {
	startWith string
	endWith   string
}

func NewStartEndLeftInner(startWith, endWith string) *startEndLeftInner {
	return &startEndLeftInner{
		startWith: startWith,
		endWith:   endWith,
	}
}

func (_ *startEndLeftInner) Name() string {
	return "start_end_left_inner"
}

func (seli *startEndLeftInner) Select(source []byte, without ...[2]int) [][2]int {
	if len(source) == 0 {
		return nil
	}

	var results [][2]int

	startRe := regexp.MustCompile(seli.startWith)
	endRe := regexp.MustCompile(seli.endWith)

	i := 0
OUTER:
	for i < len(source) {
		startLoc := startRe.FindIndex(source[i:])
		if startLoc == nil {
			break
		}
		startAbs := i + startLoc[0]

		endLoc := endRe.FindIndex(source[startAbs:])
		if endLoc == nil {
			break
		}
		contentEnd := startAbs + endLoc[0]
		endAbs := startAbs + endLoc[1]

		for _, w := range without {
			if startAbs < w[1] && contentEnd > w[0] {
				i = endAbs
				continue OUTER
			}
		}

		results = append(results, [2]int{startAbs, contentEnd})
		i = endAbs
	}

	return results
}
