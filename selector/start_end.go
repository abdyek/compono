package selector

import (
	"fmt"
	"regexp"

	"github.com/umono-cms/compono/util"
)

type startEnd struct {
	reStart *regexp.Regexp
	reEnd   *regexp.Regexp
}

func NewStartEnd(startWith, endWith string) (Selector, error) {

	reStart, err := regexp.Compile(startWith)
	if err != nil {
		return nil, fmt.Errorf("invalid start regex: %w", err)
	}

	reEnd, err := regexp.Compile(endWith)
	if err != nil {
		return nil, fmt.Errorf("invalid end regex: %w", err)
	}

	return &startEnd{
		reStart: reStart,
		reEnd:   reEnd,
	}, nil
}

func (_ *startEnd) Name() string {
	return "start_end"
}

// Select returns matched index ranges in the given source.
// The order of returned ranges is not guaranteed.
// Nested matches are ignored; only the outermost ranges are returned.
// The 'without' ranges must be provided in ascending (left-to-right) order.
func (se *startEnd) Select(source []byte, without ...[2]int) [][2]int {
	if len(source) == 0 {
		return [][2]int{}
	}

	results := [][2]int{}

	noSelected := filterNoSelected(without, len(source))
	for _, ns := range noSelected {
		results = append(results, se.slct(source, ns[0], ns[1])...)
	}

	return results
}

func (se *startEnd) slct(source []byte, start, end int) [][2]int {
	offset := start
	piece := source[start:end]

	startLocs := se.reStart.FindAllIndex(piece, -1)
	endLocs := se.reEnd.FindAllIndex(piece, -1)

	lenOfSL := len(startLocs)
	matchedEL := []int{}
	results := [][2]int{}

	for i := lenOfSL - 1; i >= 0; i-- {
		found, endIndex := se.findEL(endLocs, matchedEL, startLocs[i][1])
		if !found {
			continue
		}
		matchedEL = append(matchedEL, endIndex)
		results = append(results, [2]int{
			offset + startLocs[i][0],
			offset + endIndex,
		})
	}

	return se.eliminateNested(results)
}

func (_ *startEnd) findEL(endLocs [][]int, matchedEL []int, after int) (bool, int) {
	for _, el := range endLocs {
		if el[0] < after || util.InSliceInt(el[1], matchedEL) {
			continue
		}
		return true, el[1]
	}
	return false, 0
}

func (_ *startEnd) eliminateNested(results [][2]int) [][2]int {

	nestedInd := []int{}

	for i, r := range results {
		if !util.InSliceInt(i, nestedInd) {
			for j, rSub := range results {
				if i == j {
					continue
				}
				if rSub[0] > r[0] && rSub[1] < r[1] {
					nestedInd = append(nestedInd, j)
					continue
				}
			}
		}
	}

	eliminated := [][2]int{}
	for i, r := range results {
		if !util.InSliceInt(i, nestedInd) {
			eliminated = append(eliminated, r)
		}
	}

	return eliminated
}
