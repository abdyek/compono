package mocks

import (
	rulepkg "github.com/umono-cms/compono/rule"
	selectorpkg "github.com/umono-cms/compono/selector"
)

type rule struct {
	name      string
	selectors []selectorpkg.Selector
	rules     []rulepkg.Rule
}

func NewRule(name string, selectors []selectorpkg.Selector, rules []rulepkg.Rule) rulepkg.Rule {
	return &rule{
		name:      name,
		selectors: selectors,
		rules:     rules,
	}
}

func (r *rule) Name() string {
	return r.name
}

func (r *rule) Selectors() []selectorpkg.Selector {
	return r.selectors
}

func (r *rule) Rules() []rulepkg.Rule {
	return r.rules
}
