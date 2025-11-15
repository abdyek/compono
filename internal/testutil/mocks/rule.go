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

type ruleBuilder struct {
	name      string
	selectors []selectorpkg.Selector
	rules     []rulepkg.Rule
}

func NewRuleBuilder() *ruleBuilder {
	return &ruleBuilder{}
}

func (b *ruleBuilder) WithName(name string) *ruleBuilder {
	b.name = name
	return b
}

func (b *ruleBuilder) WithSelectors(selectors []selectorpkg.Selector) *ruleBuilder {
	b.selectors = selectors
	return b
}

func (b *ruleBuilder) WithRules(rules []rulepkg.Rule) *ruleBuilder {
	b.rules = rules
	return b
}

func (b *ruleBuilder) Build() rulepkg.Rule {
	return NewRule(b.name, b.selectors, b.rules)
}
