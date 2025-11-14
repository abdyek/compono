package selector

type startEndLeftInner struct {
	endWith  string
	startEnd *startEnd
}

func NewStartEndLeftInner(startWith, endWith string) (Selector, error) {
	se, err := NewStartEnd(startWith, endWith)
	if err != nil {
		return nil, err
	}
	return &startEndLeftInner{
		endWith:  endWith,
		startEnd: se.(*startEnd),
	}, nil
}

func (_ *startEndLeftInner) Name() string {
	return "start_end_left_inner"
}

func (seli *startEndLeftInner) Select(source []byte, without ...[2]int) [][2]int {
	selected := seli.startEnd.Select(source, without...)

	for i, index := range selected {
		matches := seli.startEnd.reEnd.FindAllIndex(source[index[0]:index[1]], -1)
		if len(matches) > 0 {
			last := matches[len(matches)-1]
			selected[i][1] = index[0] + last[0]
		}
	}

	return selected
}
