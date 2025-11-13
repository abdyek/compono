package rule

import "github.com/umono-cms/compono/selector"

type Rule interface {
	Name() string
	Selectors() []selector.Selector
	Rules() []Rule
}
