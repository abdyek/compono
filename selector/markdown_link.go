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

	for i := 0; i < len(source)-1; {
		if source[i] != ']' || source[i+1] != '(' {
			i++
			continue
		}

		start, startFound := findMarkdownLinkStart(source, i)
		if !startFound {
			i++
			continue
		}

		end, endFound := findMarkdownLinkEnd(source, i+2)
		if !endFound {
			i++
			continue
		}

		bounds = append(bounds, markdownLinkBounds{
			start:    start,
			textEnd:  i,
			urlStart: i + 2,
			end:      end,
		})
		i = end
	}

	return bounds
}

func findMarkdownLinkStart(source []byte, textEnd int) (int, bool) {
	depth := 0

	for i := textEnd - 1; i >= 0; i-- {
		switch source[i] {
		case ']':
			depth++
		case '[':
			if depth == 0 {
				return i, true
			}
			depth--
		}
	}

	return 0, false
}

func findMarkdownLinkEnd(source []byte, urlStart int) (int, bool) {
	for i := urlStart; i < len(source); i++ {
		if source[i] == ')' {
			return i + 1, true
		}
	}

	return 0, false
}

func isInsideOneOf(rng [2]int, ranges [][2]int) bool {
	for _, r := range ranges {
		if rng[0] >= r[0] && rng[1] <= r[1] {
			return true
		}
	}

	return false
}
