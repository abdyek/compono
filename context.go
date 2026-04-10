package compono

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/rule"
	"github.com/umono-cms/compono/util"
)

func buildContextWrapper(parent ast.Node, values map[string]any) (ast.Node, error) {
	wrapper := ast.DefaultEmptyNode()
	wrapper.SetRule(rule.NewDynamic("context-wrapper"))
	wrapper.SetParent(parent)

	if len(values) == 0 {
		return wrapper, nil
	}

	children := make([]ast.Node, 0, len(values))
	for key, value := range values {
		entry := ast.DefaultEmptyNode()
		entry.SetRule(rule.NewDynamic("context-entry"))
		entry.SetParent(wrapper)

		keyNode := ast.DefaultEmptyNode()
		keyNode.SetRule(rule.NewDynamic("context-key"))
		keyNode.SetParent(entry)
		keyNode.SetRaw([]byte(key))

		valueType := ast.DefaultEmptyNode()
		valueType.SetRule(rule.NewDynamic("context-value-type"))
		valueType.SetParent(entry)

		typedValue, err := buildContextValueNode(valueType, value)
		if err != nil {
			return nil, err
		}
		valueType.SetChildren([]ast.Node{typedValue})
		entry.SetChildren([]ast.Node{keyNode, valueType})
		children = append(children, entry)
	}

	wrapper.SetChildren(children)
	return wrapper, nil
}

func buildContextValueNode(parent ast.Node, value any) (ast.Node, error) {
	if value == nil {
		return nil, NewComponoError(ErrUnsupportedType, "nil is not supported in context")
	}

	val := reflect.ValueOf(value)
	if !val.IsValid() {
		return nil, NewComponoError(ErrUnsupportedType, "invalid context value")
	}

	for val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil, NewComponoError(ErrUnsupportedType, "nil is not supported in context")
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return buildContextScalarNode(parent, "comp-string-param", val.String()), nil
	case reflect.Bool:
		return buildContextScalarNode(parent, "comp-bool-param", strconv.FormatBool(val.Bool())), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return buildContextScalarNode(parent, "comp-integer-param", strconv.FormatInt(val.Int(), 10)), nil
	case reflect.Slice, reflect.Array:
		node := ast.DefaultEmptyNode()
		node.SetRule(rule.NewDynamic("comp-array-param"))
		node.SetParent(parent)

		valuesNode := ast.DefaultEmptyNode()
		valuesNode.SetRule(rule.NewDynamic("comp-array-param-values"))
		valuesNode.SetParent(node)

		children := make([]ast.Node, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			valueNode := ast.DefaultEmptyNode()
			valueNode.SetRule(rule.NewDynamic("comp-array-param-value"))
			valueNode.SetParent(valuesNode)

			valueTypeNode := ast.DefaultEmptyNode()
			valueTypeNode.SetRule(rule.NewDynamic("comp-array-param-value-type"))
			valueTypeNode.SetParent(valueNode)

			itemNode, err := buildContextValueNode(valueTypeNode, val.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			valueTypeNode.SetChildren([]ast.Node{itemNode})
			valueNode.SetChildren([]ast.Node{valueTypeNode})
			children = append(children, valueNode)
		}

		valuesNode.SetChildren(children)
		node.SetChildren([]ast.Node{valuesNode})
		return node, nil
	case reflect.Map:
		if val.Type().Key().Kind() != reflect.String {
			return nil, NewComponoError(ErrUnsupportedType, "context map keys must be strings")
		}
		return buildContextRecordNode(parent, val)
	case reflect.Struct:
		return buildContextStructNode(parent, val)
	default:
		return nil, NewComponoError(ErrUnsupportedType, fmt.Sprintf("unsupported context value type %s", val.Kind()))
	}
}

func buildContextScalarNode(parent ast.Node, ruleName string, raw string) ast.Node {
	node := ast.DefaultEmptyNode()
	node.SetRule(rule.NewDynamic(ruleName))
	node.SetParent(parent)

	def := ast.DefaultEmptyNode()
	def.SetRule(rule.NewDynamic("comp-param-defa-value"))
	def.SetParent(node)
	def.SetRaw([]byte(raw))

	node.SetChildren([]ast.Node{def})
	return node
}

func buildContextRecordNode(parent ast.Node, val reflect.Value) (ast.Node, error) {
	node := ast.DefaultEmptyNode()
	node.SetRule(rule.NewDynamic("comp-record-param"))
	node.SetParent(parent)

	valuesNode := ast.DefaultEmptyNode()
	valuesNode.SetRule(rule.NewDynamic("comp-record-param-values"))
	valuesNode.SetParent(node)

	iter := val.MapRange()
	children := make([]ast.Node, 0, val.Len())
	for iter.Next() {
		recordValueNode := ast.DefaultEmptyNode()
		recordValueNode.SetRule(rule.NewDynamic("comp-record-param-value"))
		recordValueNode.SetParent(valuesNode)

		keyNode := ast.DefaultEmptyNode()
		keyNode.SetRule(rule.NewDynamic("comp-record-param-key"))
		keyNode.SetParent(recordValueNode)
		keyNode.SetRaw([]byte(iter.Key().String()))

		valueTypeNode := ast.DefaultEmptyNode()
		valueTypeNode.SetRule(rule.NewDynamic("comp-record-param-value-type"))
		valueTypeNode.SetParent(recordValueNode)

		fieldNode, err := buildContextValueNode(valueTypeNode, iter.Value().Interface())
		if err != nil {
			return nil, err
		}
		valueTypeNode.SetChildren([]ast.Node{fieldNode})
		recordValueNode.SetChildren([]ast.Node{keyNode, valueTypeNode})
		children = append(children, recordValueNode)
	}

	valuesNode.SetChildren(children)
	node.SetChildren([]ast.Node{valuesNode})
	return node, nil
}

func buildContextStructNode(parent ast.Node, val reflect.Value) (ast.Node, error) {
	node := ast.DefaultEmptyNode()
	node.SetRule(rule.NewDynamic("comp-record-param"))
	node.SetParent(parent)

	valuesNode := ast.DefaultEmptyNode()
	valuesNode.SetRule(rule.NewDynamic("comp-record-param-values"))
	valuesNode.SetParent(node)

	children := []ast.Node{}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		key := field.Tag.Get("compono")
		if key == "" {
			key = util.ToKebabCase(field.Name)
		} else if !util.IsKebabCase(key) {
			return nil, NewComponoError(ErrUnsupportedKeyNotation, fmt.Sprintf("invalid compono struct tag %q", key))
		}

		recordValueNode := ast.DefaultEmptyNode()
		recordValueNode.SetRule(rule.NewDynamic("comp-record-param-value"))
		recordValueNode.SetParent(valuesNode)

		keyNode := ast.DefaultEmptyNode()
		keyNode.SetRule(rule.NewDynamic("comp-record-param-key"))
		keyNode.SetParent(recordValueNode)
		keyNode.SetRaw([]byte(key))

		valueTypeNode := ast.DefaultEmptyNode()
		valueTypeNode.SetRule(rule.NewDynamic("comp-record-param-value-type"))
		valueTypeNode.SetParent(recordValueNode)

		fieldNode, err := buildContextValueNode(valueTypeNode, val.Field(i).Interface())
		if err != nil {
			return nil, err
		}
		valueTypeNode.SetChildren([]ast.Node{fieldNode})
		recordValueNode.SetChildren([]ast.Node{keyNode, valueTypeNode})
		children = append(children, recordValueNode)
	}

	valuesNode.SetChildren(children)
	node.SetChildren([]ast.Node{valuesNode})
	return node, nil
}
