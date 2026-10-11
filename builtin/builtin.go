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
					IsRequired:   true,
				},
				{
					Name:         "url",
					Schema:       String(),
					DefaultValue: "",
					IsRequired:   true,
				},
				{
					Name:         "new-tab",
					Schema:       Bool(),
					DefaultValue: false,
				},
				{
					Name:         "current",
					Schema:       String(),
					DefaultValue: "",
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
					IsRequired:   true,
				},
			},
			InlineRenderable: true,
		},
		{
			Name: "HEADER",
			Params: []Param{
				{
					Name:         "content",
					Schema:       Component(),
					DefaultValue: "NO_MATTER",
					IsRequired:   true,
				},
			},
			InlineRenderable: false,
		},
		{
			Name: "MAIN",
			Params: []Param{
				{
					Name:         "content",
					Schema:       Component(),
					DefaultValue: "NO_MATTER",
					IsRequired:   true,
				},
			},
			InlineRenderable: false,
		},
		{
			Name: "FOOTER",
			Params: []Param{
				{
					Name:         "content",
					Schema:       Component(),
					DefaultValue: "NO_MATTER",
					IsRequired:   true,
				},
			},
			InlineRenderable: false,
		},
		{
			Name: "ARTICLE",
			Params: []Param{
				{
					Name:         "content",
					Schema:       Component(),
					DefaultValue: "NO_MATTER",
					IsRequired:   true,
				},
			},
			InlineRenderable: false,
		},
		{
			Name: "ASIDE",
			Params: []Param{
				{
					Name:         "content",
					Schema:       Component(),
					DefaultValue: "NO_MATTER",
					IsRequired:   true,
				},
				{
					Name:         "label",
					Schema:       String(),
					DefaultValue: "",
				},
			},
			InlineRenderable: false,
		},
		{
			Name: "NAV",
			Params: []Param{
				{
					Name:         "content",
					Schema:       Component(),
					DefaultValue: "NO_MATTER",
					IsRequired:   true,
				},
				{
					Name:         "label",
					Schema:       String(),
					DefaultValue: "",
				},
			},
			InlineRenderable: false,
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
