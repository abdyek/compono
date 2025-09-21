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

func (_ *untilFirstMatch) Name() string {
	return "until_first_math"
}

func (ufm *untilFirstMatch) Select(source []byte, without ...[2]int) [][2]int {
	re := regexp.MustCompile(ufm.regex)
	offset := 0

	for offset <= len(source) {
		loc := re.FindStringIndex(string(source[offset:]))
		if loc == nil {
			return [][2]int{{offset, len(source)}}
		}

		absEnd := offset + loc[0]
		interval := [2]int{offset, absEnd}

		conflict := false
		for _, w := range without {
			if intersects(interval[0], interval[1], w[0], w[1]) {
				conflict = true
				break
			}
		}

		if !conflict {
			return [][2]int{interval}
		}

		offset = absEnd
	}

	return nil
}
