package components

type Component interface {
	Name() string
	StartWith() string
	EndWith() string
}

func DefaultComponents() []Component {
	return []Component{
		&h1{},
	}
}
