package selector

import "regexp"

type startEndInner struct {
	startWith string
	endWith   string
}

func NewStartEndInner(startWith, endWith string) *startEndInner {
	return &startEndInner{
		startWith: startWith,
		endWith:   endWith,
	}
}

func (sei *startEndInner) Select(source []byte, without ...[2]int) [][2]int {
	var results [][2]int

	startRe := regexp.MustCompile(sei.startWith)
	endRe := regexp.MustCompile(sei.endWith)

	i := 0
OUTER:
	for i < len(source) {
		startLoc := startRe.FindIndex(source[i:])
		if startLoc == nil {
			break
		}
		contentStart := i + startLoc[1]

		endLoc := endRe.FindIndex(source[contentStart:])
		if endLoc == nil {
			break
		}
		contentEnd := contentStart + endLoc[0]
		endAbs := contentStart + endLoc[1]

		for _, w := range without {
			if contentStart < w[1] && contentEnd > w[0] {
				i = endAbs
				continue OUTER
			}
		}

		results = append(results, [2]int{contentStart, contentEnd})
		i = endAbs // endWith sonrasından devam et
	}

	return results
}
