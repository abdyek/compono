package components

type h1 struct {
}

func (h *h1) Name() string {
	return "H1"
}

func (h *h1) StartWith() string {
	return `\s*# `
}

func (h *h1) EndWith() string {
	return `\n|\z`
}
