package selector

type Selector interface {
	Select(source []byte, without ...[2]int) [][2]int
}

type Named interface {
	Name() string
}

func filterNoSelected(alreadySelected [][2]int, lenOfSrc int) [][2]int {

	lenOfAS := len(alreadySelected)

	if lenOfAS == 0 {
		return [][2]int{[2]int{0, lenOfSrc}}
	}

	noSelected := [][2]int{}

	for i := 0; i < lenOfAS-1; i++ {
		if alreadySelected[i][1] != alreadySelected[i+1][0] {
			noSelected = append(noSelected, [2]int{alreadySelected[i][1], alreadySelected[i+1][0]})
		}
	}

	if alreadySelected[0][0] != 0 {
		noSelected = append([][2]int{[2]int{0, alreadySelected[0][0]}}, noSelected...)
	}

	if alreadySelected[lenOfAS-1][1] != lenOfSrc {
		noSelected = append(noSelected, [2]int{alreadySelected[lenOfAS-1][1], lenOfSrc})
	}

	return noSelected
}
