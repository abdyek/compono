package selector

type all struct {
}

func NewAll() *all {
	return &all{}
}

func (_ *all) Name() string {
	return "all"
}

func (*all) Select(source []byte, without ...[2]int) [][2]int {
	length := len(source)
	if length == 0 {
		return nil
	}

	result := make([][2]int, 0)
	current := 0

	for _, w := range without {
		start, end := w[0], w[1]
		if current < start {
			result = append(result, [2]int{current, start})
		}
		if current < end {
			current = end
		}
	}

	if current < length {
		result = append(result, [2]int{current, length})
	}

	return result
}
