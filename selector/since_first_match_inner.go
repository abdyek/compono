package selector

import "regexp"

type sinceFirstMatchInner struct {
	regex string
}

func NewSinceFirstMatchInner(regex string) *sinceFirstMatchInner {
	return &sinceFirstMatchInner{
		regex: regex,
	}
}

func (_ *sinceFirstMatchInner) Name() string {
	return "since_first_math_inner"
}

func (sfm *sinceFirstMatchInner) Select(source []byte, without ...[2]int) [][2]int {
	re := regexp.MustCompile(sfm.regex)
	offset := 0

	for offset <= len(source) {
		loc := re.FindStringIndex(string(source[offset:]))
		if loc == nil {
			return [][2]int{}
		}

		absStart := offset + loc[0]
		interval := [2]int{offset, len(source)}

		conflict := false
		for _, w := range without {
			if intersects(interval[0], interval[1], w[0], w[1]) {
				conflict = true
				break
			}
		}

		if !conflict {
			return [][2]int{{absStart, len(source)}}
		}

		offset = absStart + 1
	}

	return nil
}
