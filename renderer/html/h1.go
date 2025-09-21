package html

type h1 struct {
}

func NewH1() *h1 {
	return &h1{}
}

func (_ *h1) Name() string {
	return "h1"
}

func (_ *h1) Void() bool {
	return false
}
