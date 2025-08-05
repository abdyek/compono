package html

type h2 struct {
}

func NewH2() *h2 {
	return &h2{}
}

func (_ *h2) Name() string {
	return "h2"
}

func (_ *h2) Void() bool {
	return false
}
