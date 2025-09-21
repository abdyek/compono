package selector

type Selector interface {
	Select(source []byte, without ...[2]int) [][2]int
}

type Named interface {
	Name() string
}
