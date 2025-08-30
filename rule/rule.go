package rule

import "github.com/umono-cms/compono/selector"

type Rule interface {
	Name() string
	Selectors() []selector.Selector
	Rules() []Rule
	scalability
}

type scalability interface {
	RewriteRules([]Rule)
}

type scalable struct {
	rules []Rule
}

func (s *scalable) RewriteRules(rules []Rule) {
	s.rules = rules
}

func OverrideRules(rules []Rule, dominantRules []Rule) []Rule {
	overridden := append([]Rule{}, rules...)

	for _, dc := range dominantRules {
		i, _ := FindRuleIndexByName(overridden, dc.Name())
		if i == -1 {
			overridden = append(overridden, dc)
		} else {
			overridden[i] = dc
		}
	}

	return overridden
}

func FindRuleIndexByName(rules []Rule, name string) (int, Rule) {
	for i, r := range rules {
		if r.Name() == name {
			return i, r
		}
	}
	return -1, nil
}
