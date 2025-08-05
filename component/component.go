package component

import "github.com/umono-cms/compono/selector"

type Component interface {
	Name() string
	Selectors() []selector.Selector
	Components() []Component
	scalability
}

type scalability interface {
	RewriteComponents([]Component)
}

type scalable struct {
	components []Component
}

func (s *scalable) RewriteComponents(components []Component) {
	s.components = components
}

func OverrideComponents(comps []Component, dominantComps []Component) []Component {
	overridden := append([]Component{}, comps...)

	for _, dc := range dominantComps {
		i, _ := FindCompIndexByName(overridden, dc.Name())
		if i == -1 {
			overridden = append(overridden, dc)
		} else {
			overridden[i] = dc
		}
	}

	return overridden
}

func FindCompIndexByName(comps []Component, name string) (int, Component) {
	for i, c := range comps {
		if c.Name() == name {
			return i, c
		}
	}
	return -1, nil
}
