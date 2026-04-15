package builtin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umono-cms/compono/ast"
)

func TestMatchesResolvedValueScalar(t *testing.T) {
	assert.True(t, MatchesResolvedValue(String(), ast.ResolvedValue{Type: "string", Raw: "hello"}))
	assert.False(t, MatchesResolvedValue(String(), ast.ResolvedValue{Type: "bool", Raw: "true"}))
}

func TestMatchesResolvedValueArray(t *testing.T) {
	schema := ArrayOf(String()).Min(1).Max(2)

	assert.True(t, MatchesResolvedValue(schema, ast.ResolvedValue{
		Type: "array",
		Items: []ast.ResolvedValue{
			{Type: "string", Raw: "a"},
			{Type: "string", Raw: "b"},
		},
	}))

	assert.False(t, MatchesResolvedValue(schema, ast.ResolvedValue{
		Type: "array",
		Items: []ast.ResolvedValue{
			{Type: "string", Raw: "a"},
			{Type: "bool", Raw: "true"},
		},
	}))
}

func TestMatchesResolvedValueHeterogeneousArray(t *testing.T) {
	schema := Array(String(), Integer(), Bool())

	assert.True(t, MatchesResolvedValue(schema, ast.ResolvedValue{
		Type: "array",
		Items: []ast.ResolvedValue{
			{Type: "string", Raw: "john"},
			{Type: "integer", Raw: "42"},
			{Type: "bool", Raw: "false"},
		},
	}))

	assert.False(t, MatchesResolvedValue(schema, ast.ResolvedValue{
		Type: "array",
		Items: []ast.ResolvedValue{
			{Type: "record", Fields: map[string]ast.ResolvedValue{"x": {Type: "string", Raw: "y"}}},
		},
	}))
}

func TestMatchesResolvedValueRecord(t *testing.T) {
	schema := Record(
		Field("title", String()).Required(),
		Field("meta", Record(
			Field("layout", Enum("hero", "stack")).Required(),
			Field("order", Integer()),
		).DisallowUnknownKeys()),
	).DisallowUnknownKeys()

	assert.True(t, MatchesResolvedValue(schema, ast.ResolvedValue{
		Type: "record",
		Fields: map[string]ast.ResolvedValue{
			"title": {Type: "string", Raw: "Hello"},
			"meta": {
				Type: "record",
				Fields: map[string]ast.ResolvedValue{
					"layout": {Type: "string", Raw: "hero"},
					"order":  {Type: "integer", Raw: "2"},
				},
			},
		},
	}))

	assert.False(t, MatchesResolvedValue(schema, ast.ResolvedValue{
		Type: "record",
		Fields: map[string]ast.ResolvedValue{
			"meta": {Type: "record", Fields: map[string]ast.ResolvedValue{}},
		},
	}))
}

func TestFindDefinitionLink(t *testing.T) {
	definition, ok := FindDefinition("LINK")
	require.True(t, ok)
	assert.Equal(t, "LINK", definition.Name)
	assert.Len(t, definition.Params, 3)
}

func TestFindDefinitionImage(t *testing.T) {
	definition, ok := FindDefinition("IMAGE")
	require.True(t, ok)
	assert.Equal(t, "IMAGE", definition.Name)
	assert.Len(t, definition.Params, 2)
}
