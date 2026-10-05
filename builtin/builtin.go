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
		{
			Name: "NAVIGATION",
			Params: []Param{
				{
					Name:         "items",
					Schema:       ArrayOf(navigationItemSchema()).Min(1),
					DefaultValue: []any{},
					IsRequired:   true,
				},
			},
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

func navigationItemSchema() ValueSchema {
	return Record(
		Field("label", String()).Required(),
		Field("target", String()).Required(),
		Field("children", ArrayOf(Lazy(navigationItemSchema))),
	).DisallowUnknownKeys()
}
