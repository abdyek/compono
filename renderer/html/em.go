package html

type em struct {
}

func NewEm() *em {
	return &em{}
}

func (_ *em) Name() string {
	return "em"
}

func (_ *em) Void() bool {
	return false
}
