package builtin

type ParamType int

const (
	StringType ParamType = iota + 1
	NumberType
	BoolType
	ComponentType
)

type Component struct {
	Name             string
	Params           []Param
	InlineRenderable bool
}

type Param struct {
	Name         string
	Type         ParamType
	DefaultValue any
}

func BuiltinComponents() []Component {
	return []Component{
		{
			Name: "LINK",
			Params: []Param{
				{
					Name:         "text",
					Type:         StringType,
					DefaultValue: "",
				},
				{
					Name:         "url",
					Type:         StringType,
					DefaultValue: "",
				},
				{
					Name:         "new-tab",
					Type:         BoolType,
					DefaultValue: false,
				},
			},
			InlineRenderable: true,
		},
	}
}
