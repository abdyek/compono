package rule

// emphasisSpans returns the [start, end) byte ranges of em (runLen=1) or
// strong (runLen=2) emphasis spans in source, with delimiters included.
// Stars inside a {{ ... }} template unit are ignored. An opener is a maximal
// '*' run whose length equals runLen and which is followed by a non-whitespace
// byte. For runLen=1, the closer is the next run; for runLen=2, it is the next
// run with length 2. Opener and closer must lie on the same line, and the
// closer must be preceded by a non-space, non-tab byte.
func emphasisSpans(source []byte, runLen int) [][2]int {
	type run struct {
		start  int
		length int
		line   int
	}

	var runs []run
	line := 1
	i := 0
	for i < len(source) {
		c := source[i]
		if c == '*' {
			start := i
			runLine := line
			for i < len(source) && source[i] == '*' {
				i++
			}
			runs = append(runs, run{start: start, length: i - start, line: runLine})
			continue
		}
		if c == '{' && i+1 < len(source) && source[i+1] == '{' {
			i += 2
			end := -1
			for j := i; j+1 < len(source); j++ {
				if source[j] == '}' && source[j+1] == '}' {
					end = j + 2
					break
				}
			}
			if end != -1 {
				for j := i; j < end-2; j++ {
					if source[j] == '\n' {
						line++
					}
				}
				i = end
			}
			continue
		}
		if c == '\n' {
			line++
		}
		i++
	}

	var spans [][2]int
	idx := 0
	for idx < len(runs) {
		r := runs[idx]
		if r.length != runLen {
			idx++
			continue
		}
		afterIdx := r.start + r.length
		if afterIdx >= len(source) {
			idx++
			continue
		}
		b := source[afterIdx]
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			idx++
			continue
		}
		var candIdx = -1
		if runLen == 1 {
			if idx+1 < len(runs) {
				candIdx = idx + 1
			}
		} else {
			for j := idx + 1; j < len(runs); j++ {
				if runs[j].length == 2 {
					candIdx = j
					break
				}
			}
		}
		if candIdx == -1 {
			idx++
			continue
		}
		cand := runs[candIdx]
		if cand.line != r.line {
			idx++
			continue
		}
		if cand.start > 0 {
			before := source[cand.start-1]
			if before == ' ' || before == '\t' {
				idx++
				continue
			}
		}
		if cand.length != runLen {
			idx++
			continue
		}
		spans = append(spans, [2]int{r.start, cand.start + cand.length})
		idx = candIdx + 1
	}
	return spans
}
