package selector

// InlineUnits returns the code spans and the {{ }} units of source, recognized
// from left to right: whichever opens first wins, and the delimiters of the
// other inside it are plain text.
func InlineUnits(source []byte) (codes [][2]int, units [][2]int) {
	i := 0
	for i < len(source) {
		switch source[i] {
		case '`':
			n := 0
			for i+n < len(source) && source[i+n] == '`' {
				n++
			}
			if n == 1 {
				start := i
				j := i + 1
				matched := false
				for j < len(source) && source[j] != '\n' {
					if source[j] == '`' {
						runLen := 0
						for j+runLen < len(source) && source[j+runLen] == '`' {
							runLen++
						}
						if runLen == 1 {
							codes = append(codes, [2]int{start, j + 1})
							i = j + 1
							matched = true
						}
						break
					}
					j++
				}
				if !matched {
					i = start + n
				}
				continue
			}
			i += n
		case '{':
			if i+1 < len(source) && source[i+1] == '{' {
				end, ok := MustacheCallEnd(source, i)
				if !ok {
					found := -1
					for k := i + 2; k+1 < len(source); k++ {
						if source[k] == '}' && source[k+1] == '}' {
							found = k + 2
							break
						}
					}
					if found == -1 {
						i += 2
						continue
					}
					end = found
				}
				units = append(units, [2]int{i, end})
				i = end
				continue
			}
			i++
		default:
			i++
		}
	}
	return
}
