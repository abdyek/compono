package builtin

import (
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
)

var webGridAreaPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z0-9]+)*$`)

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
			Name: "WEB_GRID",
			Params: []Param{
				{
					Name:         "items",
					Schema:       ArrayOf(webGridItemSchema()),
					DefaultValue: []any{},
				},
				{
					Name:         "grid-template-columns",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "grid-template-rows",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String())),
					DefaultValue: []any{},
				},
				{
					Name:         "sm-grid-template-columns",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "md-grid-template-columns",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "lg-grid-template-columns",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "xl-grid-template-columns",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "xxl-grid-template-columns",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "sm-grid-template-rows",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "md-grid-template-rows",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "lg-grid-template-rows",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "xl-grid-template-rows",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "xxl-grid-template-rows",
					Schema:       ArrayOf(String()),
					DefaultValue: []any{},
				},
				{
					Name:         "sm-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String())),
					DefaultValue: []any{},
				},
				{
					Name:         "md-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String())),
					DefaultValue: []any{},
				},
				{
					Name:         "lg-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String())),
					DefaultValue: []any{},
				},
				{
					Name:         "xl-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String())),
					DefaultValue: []any{},
				},
				{
					Name:         "xxl-grid-template-areas",
					Schema:       ArrayOf(ArrayOf(String())),
					DefaultValue: []any{},
				},
			},
		},
	}
}

func webGridItemSchema() ValueSchema {
	return Record(
		Field("component", Component()).Required(),
		Field("grid-area", String().Matching(func(value ast.ResolvedValue) bool {
			raw := strings.TrimSpace(value.Raw)
			if raw == "" {
				return false
			}
			if !webGridAreaPattern.MatchString(raw) {
				return false
			}
			for _, ch := range raw {
				if ch < '0' || ch > '9' {
					return true
				}
			}
			return false
		})).Required(),
	).DisallowUnknownKeys()
}
