package selector

type bounds struct {
	start Selector
	end   Selector
}

func NewBounds(start, end Selector) Selector {
	return &bounds{
		start: start,
		end:   end,
	}
}

func (b *bounds) Name() string {
	startName := "unknown"
	endName := "unknown"

	if s, ok := b.start.(Named); ok {
		startName = s.Name()
	}

	if s, ok := b.end.(Named); ok {
		endName = s.Name()
	}

	return "bounds(" + startName + ", " + endName + ")"
}

func (b *bounds) Select(source []byte, without ...[2]int) [][2]int {
	if len(source) == 0 {
		return [][2]int{}
	}

	noSelected := filterNoSelected(without, len(source))

	startIndexes := b.start.Select(source, without...)
	endIndexes := b.end.Select(source, without...)

	results := [][2]int{}

	for _, ns := range noSelected {
		minInd := b.minStartIndex(startIndexes, ns[0], ns[1])
		maxInd := b.maxEndIndex(endIndexes, ns[0], ns[1])

		if minInd == -1 || maxInd == -1 || minInd > maxInd {
			continue
		}

		results = append(results, [2]int{minInd, maxInd})
	}
	return results
}

func (b *bounds) minStartIndex(indexes [][2]int, start, end int) int {
	minInd := -1
	for _, i := range indexes {
		if i[0] < start || i[1] > end {
			continue
		}

		if minInd == -1 {
			minInd = i[0]
			continue
		}

		if minInd > i[0] {
			minInd = i[0]
		}
	}
	return minInd
}

func (b *bounds) maxEndIndex(indexes [][2]int, start, end int) int {
	maxInd := -1
	for _, i := range indexes {
		if i[0] < start || i[1] > end {
			continue
		}

		if maxInd == -1 {
			maxInd = i[1]
			continue
		}

		if maxInd < i[1] {
			maxInd = i[1]
		}
	}
	return maxInd
}
