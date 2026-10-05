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
		{
			Name: "IMAGE",
			Params: []Param{
				{
					Name:         "media",
					Schema:       imageMediaSchema(),
					DefaultValue: map[string]any{},
					IsRequired:   true,
				},
				{
					Name:         "alt",
					Schema:       String(),
					DefaultValue: "",
				},
			},
			InlineRenderable: true,
		},
	}
}

func imageMediaSchema() ValueSchema {
	return Record(
		Field("url", String()).Required(),
		Field("width", Integer()).Required(),
		Field("height", Integer()).Required(),
		Field("mime-type", String()).Required(),
		Field("variants", ArrayOf(imageVariantSchema())),
	).DisallowUnknownKeys()
}

func imageVariantSchema() ValueSchema {
	return Record(
		Field("url", String()).Required(),
		Field("width", Integer()).Required(),
		Field("height", Integer()).Required(),
		Field("mime-type", String()).Required(),
	).DisallowUnknownKeys()
}
