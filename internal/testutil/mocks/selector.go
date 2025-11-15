package mocks

import selectorpkg "github.com/umono-cms/compono/selector"

type selector struct {
	handler func([]byte, ...[2]int) [][2]int
}

func NewSelector(handler func([]byte, ...[2]int) [][2]int) selectorpkg.Selector {
	return &selector{
		handler: handler,
	}
}

func (s *selector) Select(source []byte, without ...[2]int) [][2]int {
	return s.handler(source, without...)
}
