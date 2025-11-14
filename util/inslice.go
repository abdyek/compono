package util

func InSliceString(needle string, haystack []string) bool {
	for _, el := range haystack {
		if el == needle {
			return true
		}
	}
	return false
}

func InSliceInt(needle int, haystack []int) bool {
	for _, el := range haystack {
		if el == needle {
			return true
		}
	}
	return false
}
