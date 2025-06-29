package selector

import (
	"regexp"
)

type startEnd struct {
	startWith string
	endWith   string
}

func NewStartEnd(startWith, endWith string) *startEnd {
	return &startEnd{
		startWith: startWith,
		endWith:   endWith,
	}
}

func (se *startEnd) Select(source []byte, without ...[2]int) [][2]int {

	startRe := regexp.MustCompile(se.startWith)
	endRe := regexp.MustCompile(se.endWith)

	var results [][2]int
	offset := 0

	for {
		startLoc := startRe.FindIndex(source[offset:])
		if startLoc == nil {
			break
		}
		startAbs := offset + startLoc[0]
		searchAfterStart := offset + startLoc[1]

		endLoc := endRe.FindIndex(source[searchAfterStart:])
		if endLoc == nil {
			break
		}
		endAbs := searchAfterStart + endLoc[1]

		results = append(results, [2]int{startAbs, endAbs})

		offset = endAbs
	}

	return results
}
