package builtin

func BuiltinComponents() []Definition {
	return []Definition{
		{
			Name: "LINK",
			Params: []Param{
				{
					Name:         "text",
					Schema:       String(),
					DefaultValue: "",
				},
				{
					Name:         "url",
					Schema:       String(),
					DefaultValue: "",
				},
				{
					Name:         "new-tab",
					Schema:       Bool(),
					DefaultValue: false,
				},
			},
			InlineRenderable: true,
		},
	}
}
