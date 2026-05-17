package builtin

import (
	"strconv"

	"github.com/umono-cms/compono/ast"
)

func FindDefinition(name string) (Definition, bool) {
	for _, definition := range BuiltinComponents() {
		if definition.Name == name {
			return definition, true
		}
	}

	return Definition{}, false
}

func MatchesResolvedValue(schema ValueSchema, value ast.ResolvedValue) bool {
	switch typed := schema.(type) {
	case *ScalarSchema:
		return matchesScalarSchema(typed, value)
	case *ArraySchema:
		return matchesArraySchema(typed, value)
	case *RecordSchema:
		return matchesRecordSchema(typed, value)
	case *LazySchema:
		if typed == nil || typed.Resolver == nil {
			return false
		}
		return MatchesResolvedValue(typed.Resolver(), value)
	default:
		return false
	}
}

func matchesScalarSchema(schema *ScalarSchema, value ast.ResolvedValue) bool {
	if kindFromResolvedValue(value) != schema.Kind() {
		return false
	}

	if len(schema.AllowedValues) == 0 {
		if schema.Matcher != nil {
			return schema.Matcher(value)
		}
		return true
	}

	normalized, ok := normalizeScalarResolvedValue(value)
	if !ok {
		return false
	}

	for _, allowed := range schema.AllowedValues {
		if scalarValuesEqual(normalized, allowed) {
			if schema.Matcher != nil {
				return schema.Matcher(value)
			}
			return true
		}
	}

	return false
}

func matchesArraySchema(schema *ArraySchema, value ast.ResolvedValue) bool {
	if value.Type != "array" {
		return false
	}

	if schema.MinLen >= 0 && len(value.Items) < schema.MinLen {
		return false
	}
	if schema.MaxLen >= 0 && len(value.Items) > schema.MaxLen {
		return false
	}

	switch schema.ElementsMode {
	case ArrayElementsTuple:
		if len(schema.ElementSchemas) != len(value.Items) {
			return false
		}
		for i, item := range value.Items {
			if !MatchesResolvedValue(schema.ElementSchemas[i], item) {
				return false
			}
		}
		return true
	case ArrayElementsAnyOf:
		for _, item := range value.Items {
			matched := false
			for _, candidate := range schema.ElementSchemas {
				if MatchesResolvedValue(candidate, item) {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
		return true
	default:
		if len(schema.ElementSchemas) == 0 {
			return true
		}
		for _, item := range value.Items {
			if !MatchesResolvedValue(schema.ElementSchemas[0], item) {
				return false
			}
		}
		return true
	}
}

func matchesRecordSchema(schema *RecordSchema, value ast.ResolvedValue) bool {
	if value.Type != "record" {
		return false
	}

	if schema.MinFields >= 0 && len(value.Fields) < schema.MinFields {
		return false
	}
	if schema.MaxFields >= 0 && len(value.Fields) > schema.MaxFields {
		return false
	}

	fieldByName := make(map[string]RecordField, len(schema.Fields))
	for _, field := range schema.Fields {
		fieldByName[field.Name] = field
	}

	if !schema.AllowUnknownKeys {
		for key := range value.Fields {
			if _, ok := fieldByName[key]; !ok {
				return false
			}
		}
	}

	for _, field := range schema.Fields {
		fieldValue, ok := value.Fields[field.Name]
		if !ok {
			if field.IsRequired {
				return false
			}
			continue
		}

		if !MatchesResolvedValue(field.Schema, fieldValue) {
			return false
		}
	}

	return true
}

func kindFromResolvedValue(value ast.ResolvedValue) ValueKind {
	switch value.Type {
	case "string":
		return StringKind
	case "integer":
		return IntegerKind
	case "bool":
		return BoolKind
	case "comp":
		return ComponentKind
	case "array":
		return ArrayKind
	case "record":
		return RecordKind
	default:
		return 0
	}
}

func normalizeScalarResolvedValue(value ast.ResolvedValue) (any, bool) {
	switch value.Type {
	case "string", "comp":
		return value.Raw, true
	case "integer":
		integer, err := strconv.ParseInt(value.Raw, 10, 64)
		if err != nil {
			return nil, false
		}
		return integer, true
	case "bool":
		boolean, err := strconv.ParseBool(value.Raw)
		if err != nil {
			return nil, false
		}
		return boolean, true
	default:
		return nil, false
	}
}

func scalarValuesEqual(left any, right any) bool {
	leftInteger, leftIsInteger := integerValue(left)
	rightInteger, rightIsInteger := integerValue(right)
	if leftIsInteger || rightIsInteger {
		return leftIsInteger && rightIsInteger && leftInteger == rightInteger
	}
	return left == right
}

func integerValue(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int8:
		return int64(typed), true
	case int16:
		return int64(typed), true
	case int32:
		return int64(typed), true
	case int64:
		return typed, true
	case uint:
		return int64(typed), true
	case uint8:
		return int64(typed), true
	case uint16:
		return int64(typed), true
	case uint32:
		return int64(typed), true
	case uint64:
		return int64(typed), true
	default:
		return 0, false
	}
}

func scalarKindOf(value any) (ValueKind, bool) {
	switch value.(type) {
	case string:
		return StringKind, true
	case bool:
		return BoolKind, true
	}

	if _, ok := integerValue(value); ok {
		return IntegerKind, true
	}

	return 0, false
}
