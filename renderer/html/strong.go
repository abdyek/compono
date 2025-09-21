package html

type strong struct {
}

func NewStrong() *strong {
	return &strong{}
}

func (_ *strong) Name() string {
	return "strong"
}

func (_ *strong) Void() bool {
	return false
}
