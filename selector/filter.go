package selector

type filter struct {
	selector Selector
	callback func([]byte, [][2]int) [][2]int
}

func NewFilter(selector Selector, callback func([]byte, [][2]int) [][2]int) Selector {
	return &filter{
		selector: selector,
		callback: callback,
	}
}

func (f *filter) Name() string {
	return "filter"
}

func (f *filter) Select(source []byte, without ...[2]int) [][2]int {
	indexes := f.selector.Select(source, without...)
	return f.callback(source, indexes)
}
