package selector

type markdownLinkPart int

const (
	MarkdownLinkWhole markdownLinkPart = iota
	MarkdownLinkText
	MarkdownLinkURL
)

type markdownLinkBounds struct {
	start    int
	textEnd  int
	urlStart int
	end      int
}

type markdownLink struct {
	part markdownLinkPart
}

func NewMarkdownLink(part markdownLinkPart) Selector {
	return &markdownLink{
		part: part,
	}
}

func (_ *markdownLink) Name() string {
	return "markdown_link"
}

func (ml *markdownLink) Select(source []byte, without ...[2]int) [][2]int {
	results := [][2]int{}

	if len(source) == 0 {
		return results
	}

	noSelected := filterNoSelected(without, len(source))

	for _, bounds := range scanMarkdownLinks(source) {
		rng := ml.rangeOfPart(bounds)
		if !isInsideOneOf(rng, noSelected) {
			continue
		}
		results = append(results, rng)
	}

	return results
}

func (ml *markdownLink) rangeOfPart(bounds markdownLinkBounds) [2]int {
	switch ml.part {
	case MarkdownLinkText:
		return [2]int{bounds.start + 1, bounds.textEnd}
	case MarkdownLinkURL:
		return [2]int{bounds.urlStart, bounds.end - 1}
	default:
		return [2]int{bounds.start, bounds.end}
	}
}

func scanMarkdownLinks(source []byte) []markdownLinkBounds {
	bounds := []markdownLinkBounds{}

	codes, units := InlineUnits(source)
	ranges := append([]([2]int){}, codes...)
	ranges = append(ranges, units...)

	for i := 0; i < len(source); i++ {
		if source[i] != '[' || isInsideRanges(i, ranges) {
			continue
		}

		depth := 0
		textEnd := -1
		failed := false
		for j := i + 1; j < len(source); j++ {
			if isInsideRanges(j, ranges) {
				continue
			}
			switch source[j] {
			case '\n':
				failed = true
				j = len(source)
			case '[':
				depth++
			case ']':
				if depth == 0 {
					textEnd = j
					break
				}
				depth--
			}
			if textEnd != -1 {
				break
			}
		}

		if failed || textEnd == -1 {
			continue
		}

		if textEnd+1 >= len(source) || source[textEnd+1] != '(' || isInsideRanges(textEnd+1, ranges) {
			continue
		}

		closingParen := -1
		for k := textEnd + 2; k < len(source); k++ {
			if isInsideRanges(k, ranges) {
				continue
			}
			if source[k] == '\n' {
				break
			}
			if source[k] == ')' {
				closingParen = k
				break
			}
		}

		if closingParen == -1 {
			continue
		}

		bounds = append(bounds, markdownLinkBounds{
			start:    i,
			textEnd:  textEnd,
			urlStart: textEnd + 2,
			end:      closingParen + 1,
		})
		i = closingParen
	}

	return bounds
}

func isInsideRanges(pos int, ranges [][2]int) bool {
	for _, r := range ranges {
		if pos >= r[0] && pos < r[1] {
			return true
		}
	}
	return false
}

func isInsideOneOf(rng [2]int, ranges [][2]int) bool {
	for _, r := range ranges {
		if rng[0] >= r[0] && rng[1] <= r[1] {
			return true
		}
	}

	return false
}
