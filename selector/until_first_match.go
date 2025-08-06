package selector

import "regexp"

type untilFirstMatch struct {
	regex string
}

func NewUntilFirstMatch(regex string) *untilFirstMatch {
	return &untilFirstMatch{
		regex: regex,
	}
}

func (ufm *untilFirstMatch) Select(source []byte, without ...[2]int) [][2]int {

	// TODO: add without parameter

	re := regexp.MustCompile(ufm.regex)
	loc := re.FindStringIndex(string(source))

	if loc == nil {
		return [][2]int{
			[2]int{0, len(source)},
		}
	}

	return [][2]int{
		[2]int{0, loc[0]},
	}
}
