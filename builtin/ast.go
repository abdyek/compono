package builtin

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
)

func BuildASTNodes(parent ast.Node) []ast.Node {
	builtinComps := BuiltinComponents()
	if len(builtinComps) == 0 {
		return nil
	}

	children := make([]ast.Node, 0, len(builtinComps))

	for _, comp := range builtinComps {
		builtinComp := ast.DefaultEmptyNode()
		builtinComp.SetRule(rule.NewDynamic("builtin-comp"))
		builtinComp.SetParent(parent)

		builtinCompName := ast.DefaultEmptyNode()
		builtinCompName.SetRule(rule.NewDynamic("builtin-comp-name"))
		builtinCompName.SetParent(builtinComp)
		builtinCompName.SetRaw([]byte(comp.Name))

		compParams := ast.DefaultEmptyNode()
		compParams.SetRule(rule.NewDynamic("comp-params"))
		compParams.SetParent(builtinComp)
		compParams.SetChildren(makeBuiltinCompParams(compParams, comp.Params))

		builtinComp.SetChildren([]ast.Node{builtinCompName, compParams})
		children = append(children, builtinComp)
	}

	return children
}

func makeBuiltinCompParams(parent ast.Node, params []Param) []ast.Node {
	if len(params) == 0 {
		return nil
	}

	children := make([]ast.Node, 0, len(params))
	for _, param := range params {
		compParam := ast.DefaultEmptyNode()
		compParam.SetRule(rule.NewDynamic("comp-param"))
		compParam.SetParent(parent)

		compParamName := ast.DefaultEmptyNode()
		compParamName.SetRule(rule.NewDynamic("comp-param-name"))
		compParamName.SetParent(compParam)
		compParamName.SetRaw([]byte(param.Name))

		compParamType := ast.DefaultEmptyNode()
		compParamType.SetRule(rule.NewDynamic("comp-param-type"))
		compParamType.SetParent(compParam)

		typedParam := makeSchemaValueNode(compParamType, param.Schema, param.DefaultValue)
		if typedParam != nil {
			compParamType.SetChildren([]ast.Node{typedParam})
		}

		compParam.SetChildren([]ast.Node{compParamName, compParamType})
		children = append(children, compParam)
	}

	return children
}

func makeSchemaValueNode(parent ast.Node, schema ValueSchema, value any) ast.Node {
	if schema == nil {
		return nil
	}

	switch typed := schema.(type) {
	case *ScalarSchema:
		return makeScalarNode(parent, typed, value)
	case *ArraySchema:
		return makeArrayNode(parent, typed, value)
	case *RecordSchema:
		return makeRecordNode(parent, typed, value)
	case *LazySchema:
		if typed == nil || typed.Resolver == nil {
			return nil
		}
		return makeSchemaValueNode(parent, typed.Resolver(), value)
	default:
		return nil
	}
}

func makeScalarNode(parent ast.Node, schema *ScalarSchema, value any) ast.Node {
	typedParam := ast.DefaultEmptyNode()
	typedParam.SetRule(rule.NewDynamic(schemaRuleName(schema.Kind())))
	typedParam.SetParent(parent)

	defaultValue := ast.DefaultEmptyNode()
	defaultValue.SetRule(rule.NewDynamic("comp-param-defa-value"))
	defaultValue.SetParent(typedParam)
	defaultValue.SetRaw([]byte(formatScalarValue(schema.Kind(), value)))

	typedParam.SetChildren([]ast.Node{defaultValue})
	return typedParam
}

func makeArrayNode(parent ast.Node, schema *ArraySchema, value any) ast.Node {
	typedParam := ast.DefaultEmptyNode()
	typedParam.SetRule(rule.NewDynamic("comp-array-param"))
	typedParam.SetParent(parent)

	valuesNode := ast.DefaultEmptyNode()
	valuesNode.SetRule(rule.NewDynamic("comp-array-param-values"))
	valuesNode.SetParent(typedParam)

	items, _ := normalizeArrayLikeForAST(value)
	valuesChildren := make([]ast.Node, 0, len(items))
	elementSchema := arrayElementSchemaForAST(schema)
	for _, item := range items {
		valueNode := ast.DefaultEmptyNode()
		valueNode.SetRule(rule.NewDynamic("comp-array-param-value"))
		valueNode.SetParent(valuesNode)

		valueTypeNode := ast.DefaultEmptyNode()
		valueTypeNode.SetRule(rule.NewDynamic("comp-array-param-value-type"))
		valueTypeNode.SetParent(valueNode)

		nested := makeSchemaValueNode(valueTypeNode, schemaForNestedValue(elementSchema, item), item)
		if nested != nil {
			valueTypeNode.SetChildren([]ast.Node{nested})
		}

		valueNode.SetChildren([]ast.Node{valueTypeNode})
		valuesChildren = append(valuesChildren, valueNode)
	}

	valuesNode.SetChildren(valuesChildren)
	typedParam.SetChildren([]ast.Node{valuesNode})
	return typedParam
}

func makeRecordNode(parent ast.Node, schema *RecordSchema, value any) ast.Node {
	typedParam := ast.DefaultEmptyNode()
	typedParam.SetRule(rule.NewDynamic("comp-record-param"))
	typedParam.SetParent(parent)

	valuesNode := ast.DefaultEmptyNode()
	valuesNode.SetRule(rule.NewDynamic("comp-record-param-values"))
	valuesNode.SetParent(typedParam)

	fields, _ := normalizeRecordLikeForAST(value)
	keys := orderedRecordKeys(schema, fields)
	valuesChildren := make([]ast.Node, 0, len(keys))
	for _, key := range keys {
		fieldValue := fields[key]

		valueNode := ast.DefaultEmptyNode()
		valueNode.SetRule(rule.NewDynamic("comp-record-param-value"))
		valueNode.SetParent(valuesNode)

		keyNode := ast.DefaultEmptyNode()
		keyNode.SetRule(rule.NewDynamic("comp-record-param-key"))
		keyNode.SetParent(valueNode)
		keyNode.SetRaw([]byte(key))

		valueTypeNode := ast.DefaultEmptyNode()
		valueTypeNode.SetRule(rule.NewDynamic("comp-record-param-value-type"))
		valueTypeNode.SetParent(valueNode)

		nested := makeSchemaValueNode(valueTypeNode, recordFieldSchemaForAST(schema, key, fieldValue), fieldValue)
		if nested != nil {
			valueTypeNode.SetChildren([]ast.Node{nested})
		}

		valueNode.SetChildren([]ast.Node{keyNode, valueTypeNode})
		valuesChildren = append(valuesChildren, valueNode)
	}

	valuesNode.SetChildren(valuesChildren)
	typedParam.SetChildren([]ast.Node{valuesNode})
	return typedParam
}

func schemaRuleName(kind ValueKind) string {
	switch kind {
	case StringKind:
		return "comp-string-param"
	case IntegerKind:
		return "comp-integer-param"
	case BoolKind:
		return "comp-bool-param"
	case ComponentKind:
		return "comp-comp-param"
	case ArrayKind:
		return "comp-array-param"
	case RecordKind:
		return "comp-record-param"
	default:
		return "comp-string-param"
	}
}

func formatScalarValue(kind ValueKind, value any) string {
	if resolved, ok := value.(ast.ResolvedValue); ok {
		return resolved.Raw
	}

	switch kind {
	case StringKind, ComponentKind:
		if value == nil {
			return ""
		}
		return fmt.Sprint(value)
	case IntegerKind:
		switch typed := value.(type) {
		case int:
			return strconv.Itoa(typed)
		case int8:
			return strconv.FormatInt(int64(typed), 10)
		case int16:
			return strconv.FormatInt(int64(typed), 10)
		case int32:
			return strconv.FormatInt(int64(typed), 10)
		case int64:
			return strconv.FormatInt(typed, 10)
		case uint:
			return strconv.FormatUint(uint64(typed), 10)
		case uint8:
			return strconv.FormatUint(uint64(typed), 10)
		case uint16:
			return strconv.FormatUint(uint64(typed), 10)
		case uint32:
			return strconv.FormatUint(uint64(typed), 10)
		case uint64:
			return strconv.FormatUint(typed, 10)
		default:
			return fmt.Sprint(value)
		}
	case BoolKind:
		if typed, ok := value.(bool); ok {
			return strconv.FormatBool(typed)
		}
		return fmt.Sprint(value)
	default:
		return fmt.Sprint(value)
	}
}

func arrayElementSchemaForAST(schema *ArraySchema) ValueSchema {
	if schema == nil || len(schema.ElementSchemas) == 0 {
		return String()
	}

	switch schema.ElementsMode {
	case ArrayElementsTuple:
		return nil
	default:
		return schema.ElementSchemas[0]
	}
}

func schemaForNestedValue(fallback ValueSchema, value any) ValueSchema {
	if fallback != nil {
		return fallback
	}

	if resolved, ok := value.(ast.ResolvedValue); ok {
		switch resolved.Type {
		case "string":
			return String()
		case "integer":
			return Integer()
		case "bool":
			return Bool()
		case "comp":
			return Component()
		case "array":
			return Array()
		case "record":
			return Record()
		}
	}

	if kind, ok := scalarKindOf(value); ok {
		switch kind {
		case StringKind:
			return String()
		case IntegerKind:
			return Integer()
		case BoolKind:
			return Bool()
		case ComponentKind:
			return Component()
		}
	}

	if _, ok := normalizeArrayLikeForAST(value); ok {
		return Array()
	}
	if _, ok := normalizeRecordLikeForAST(value); ok {
		return Record()
	}

	return String()
}

func recordFieldSchemaForAST(schema *RecordSchema, key string, value any) ValueSchema {
	if schema != nil {
		for _, field := range schema.Fields {
			if field.Name == key && field.Schema != nil {
				return field.Schema
			}
		}
	}
	return schemaForNestedValue(nil, value)
}

func orderedRecordKeys(schema *RecordSchema, fields map[string]any) []string {
	keys := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))

	if schema != nil {
		for _, field := range schema.Fields {
			if _, ok := fields[field.Name]; ok {
				keys = append(keys, field.Name)
				seen[field.Name] = struct{}{}
			}
		}
	}

	extraKeys := make([]string, 0, len(fields)-len(keys))
	for key := range fields {
		if _, ok := seen[key]; ok {
			continue
		}
		extraKeys = append(extraKeys, key)
	}
	sort.Strings(extraKeys)

	return append(keys, extraKeys...)
}

func normalizeArrayLikeForAST(value any) ([]any, bool) {
	if value == nil {
		return []any{}, true
	}

	if resolved, ok := value.(ast.ResolvedValue); ok {
		if resolved.Type != "array" {
			return nil, false
		}

		items := make([]any, 0, len(resolved.Items))
		for _, item := range resolved.Items {
			items = append(items, item)
		}
		return items, true
	}

	val := reflect.ValueOf(value)
	if !val.IsValid() {
		return []any{}, true
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, false
	}

	items := make([]any, 0, val.Len())
	for i := 0; i < val.Len(); i++ {
		items = append(items, val.Index(i).Interface())
	}
	return items, true
}

func normalizeRecordLikeForAST(value any) (map[string]any, bool) {
	if value == nil {
		return map[string]any{}, true
	}

	if resolved, ok := value.(ast.ResolvedValue); ok {
		if resolved.Type != "record" {
			return nil, false
		}

		fields := make(map[string]any, len(resolved.Fields))
		for key, fieldValue := range resolved.Fields {
			fields[key] = fieldValue
		}
		return fields, true
	}

	val := reflect.ValueOf(value)
	if !val.IsValid() {
		return map[string]any{}, true
	}
	if val.Kind() != reflect.Map || val.Type().Key().Kind() != reflect.String {
		return nil, false
	}

	fields := make(map[string]any, val.Len())
	iter := val.MapRange()
	for iter.Next() {
		fields[iter.Key().String()] = iter.Value().Interface()
	}
	return fields, true
}
