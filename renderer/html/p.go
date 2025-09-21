package html

type p struct {
}

func NewP() *p {
	return &p{}
}

func (_ *p) Name() string {
	return "p"
}

func (_ *p) Void() bool {
	return false
}
