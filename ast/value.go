package ast

import "strings"

type ResolvedValue struct {
	Type              string
	Raw               string
	Items             []ResolvedValue
	Fields            map[string]ResolvedValue
	Scope             Node
	MissingContextKey string
}

func (rv ResolvedValue) IsZero() bool {
	return rv.Type == "" && rv.Raw == "" && len(rv.Items) == 0 && len(rv.Fields) == 0 && rv.Scope == nil && rv.MissingContextKey == ""
}

type AccessErrorKind string

const (
	AccessErrorUnknownRecordKey     AccessErrorKind = "unknown_record_key"
	AccessErrorArrayIndexOutOfRange AccessErrorKind = "array_index_out_of_range"
	AccessErrorInvalidKeyAccess     AccessErrorKind = "invalid_key_access"
	AccessErrorInvalidIndexAccess   AccessErrorKind = "invalid_index_access"
)

type AccessError struct {
	Kind AccessErrorKind
	Key  string
}

func ResolveCompCallArgValue(root Node, compCallArg Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	compCallArgType := FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	resolved := resolveValueNode(root, compCallArgType, invokerAncestors, currentCompCall)
	if resolved.Scope == nil {
		resolved.Scope = GetLocalCompSourceFromNode(currentCompCall, root)
	}
	return resolved
}

func ResolveParamFromAncestors(root Node, paramName string, accessors []ValueAccessor, invokerAncestors []Node) ResolvedValue {
	for _, anc := range invokerAncestors {
		if !IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call", "param-ref"}) {
			continue
		}

		compCallArgs := FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := FindNode(compCallArgs.Children(), func(cca Node) bool {
				return GetArgNameFromCompCallArg(cca) == paramName
			})

			if compCallArg != nil {
				return ApplyAccessors(ResolveCompCallArgValue(root, compCallArg, invokerAncestors, anc), accessors)
			}
		}

		if IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call"}) {
			if resolved := ResolveParamDefaultFromCompCall(root, anc, paramName); !resolved.IsZero() {
				return ApplyAccessors(resolved, accessors)
			}
		}
	}

	return ResolvedValue{}
}

func ResolveParamDefaultFromCompCall(root Node, compCallNode Node, paramName string) ResolvedValue {
	compDef := FindCompDef(root, compCallNode, getCompCallName(compCallNode))
	if compDef == nil {
		return ResolvedValue{}
	}

	return ResolveCompParamDefaultFromCompDef(root, compDef, paramName)
}

func ResolveCompParamDefaultFromCompDef(root Node, compDef Node, paramName string) ResolvedValue {
	compParam := FindNode(GetCompParamsFromCompDef(compDef), func(cp Node) bool {
		return GetParamNameFromCompParam(cp) == paramName
	})
	if compParam == nil {
		return ResolvedValue{}
	}

	resolved := resolveValueNode(root, FindNodeByRuleName(compParam.Children(), "comp-param-type"), nil, compDef)
	if resolved.Scope == nil {
		resolved.Scope = GetLocalCompSourceFromNode(compDef, root)
	}
	return resolved
}

func resolveValueNode(root Node, node Node, invokerAncestors []Node, currentNode Node) ResolvedValue {
	if node == nil {
		return ResolvedValue{}
	}

	if IsRuleNameOneOf(node, []string{
		"comp-param-type",
		"comp-call-arg-type",
		"comp-array-param-value-type",
		"comp-call-array-arg-value-type",
		"comp-record-param-value-type",
		"comp-call-record-arg-value-type",
		"context-value-type",
	}) {
		child := firstTypedValueNode(node.Children())
		if child == nil {
			return ResolvedValue{}
		}
		return resolveValueNode(root, child, invokerAncestors, currentNode)
	}

	switch node.Rule().Name() {
	case "comp-context-param", "comp-call-context-arg", "context-ref":
		return ResolveContextReferenceValue(root, node)
	case "comp-string-param", "comp-integer-param", "comp-bool-param", "comp-comp-param":
		value := FindNodeByRuleName(node.Children(), "comp-param-defa-value")
		raw := strings.TrimSpace(string(node.Raw()))
		if value != nil {
			raw = strings.TrimSpace(string(value.Raw()))
		}
		return ResolvedValue{
			Type: strings.TrimSuffix(strings.TrimPrefix(node.Rule().Name(), "comp-"), "-param"),
			Raw:  raw,
		}
	case "comp-call-string-arg", "comp-call-integer-arg", "comp-call-bool-arg", "comp-call-comp-arg":
		value := FindNodeByRuleName(node.Children(), "comp-call-arg-value")
		raw := strings.TrimSpace(string(node.Raw()))
		if value != nil {
			raw = strings.TrimSpace(string(value.Raw()))
		}
		return ResolvedValue{
			Type: strings.TrimSuffix(strings.TrimPrefix(node.Rule().Name(), "comp-call-"), "-arg"),
			Raw:  raw,
		}
	case "comp-call-param-arg":
		argValue := FindNodeByRuleName(node.Children(), "comp-call-arg-value")
		if argValue == nil {
			return ResolvedValue{}
		}

		referencedParamName, accessors := GetValuePathFromRaw(strings.TrimSpace(string(argValue.Raw())))

		remainingAncestors := invokerAncestors
		for i, anc := range invokerAncestors {
			if anc == currentNode {
				remainingAncestors = invokerAncestors[i+1:]
				break
			}
		}

		return ResolveParamFromAncestors(root, referencedParamName, accessors, remainingAncestors)
	case "comp-array-param":
		return resolveArrayValues(node, "comp-array-param-values", "comp-array-param-value", "comp-array-param-value-type", root, invokerAncestors, currentNode)
	case "comp-record-param":
		return resolveRecordValues(node, "comp-record-param-values", "comp-record-param-value", "comp-record-param-key", "comp-record-param-value-type", root, invokerAncestors, currentNode)
	case "comp-call-array-arg":
		return resolveArrayValues(node, "comp-call-array-arg-values", "comp-call-array-arg-value", "comp-call-array-arg-value-type", root, invokerAncestors, currentNode)
	case "comp-call-record-arg":
		return resolveRecordValues(node, "comp-call-record-arg-values", "comp-call-record-arg-value", "comp-call-record-arg-key", "comp-call-record-arg-value-type", root, invokerAncestors, currentNode)
	default:
		if len(node.Children()) == 1 {
			return resolveValueNode(root, node.Children()[0], invokerAncestors, currentNode)
		}
	}

	return ResolvedValue{}
}

func resolveArrayValues(node Node, valuesRuleName, valueRuleName, valueTypeRuleName string, root Node, invokerAncestors []Node, currentNode Node) ResolvedValue {
	values := FindNodeByRuleName(node.Children(), valuesRuleName)
	if values == nil {
		return ResolvedValue{Type: "array", Items: []ResolvedValue{}}
	}

	items := []ResolvedValue{}
	for _, valueNode := range values.Children() {
		if !IsRuleName(valueNode, valueRuleName) {
			continue
		}
		resolved := resolveValueNode(root, FindNodeByRuleName(valueNode.Children(), valueTypeRuleName), invokerAncestors, currentNode)
		if resolved.IsZero() {
			continue
		}
		items = append(items, resolved)
	}

	return ResolvedValue{
		Type:  "array",
		Items: items,
	}
}

func resolveRecordValues(node Node, valuesRuleName, valueRuleName, keyRuleName, valueTypeRuleName string, root Node, invokerAncestors []Node, currentNode Node) ResolvedValue {
	values := FindNodeByRuleName(node.Children(), valuesRuleName)
	if values == nil {
		return ResolvedValue{Type: "record", Fields: map[string]ResolvedValue{}}
	}

	fields := map[string]ResolvedValue{}
	for _, valueNode := range values.Children() {
		if !IsRuleName(valueNode, valueRuleName) {
			continue
		}

		keyNode := FindNodeByRuleName(valueNode.Children(), keyRuleName)
		valueTypeNode := FindNodeByRuleName(valueNode.Children(), valueTypeRuleName)
		if keyNode == nil || valueTypeNode == nil {
			continue
		}

		resolved := resolveValueNode(root, valueTypeNode, invokerAncestors, currentNode)
		if resolved.IsZero() {
			continue
		}
		fields[GetParamRefName(keyNode)] = resolved
	}

	return ResolvedValue{
		Type:   "record",
		Fields: fields,
	}
}

func resolveCompCallArrayArgValue(root Node, node Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	return resolveArrayValues(node, "comp-call-array-arg-values", "comp-call-array-arg-value", "comp-call-array-arg-value-type", root, invokerAncestors, currentCompCall)
}

func resolveCompCallArrayValueType(root Node, node Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	return resolveValueNode(root, node, invokerAncestors, currentCompCall)
}

func resolveCompCallRecordArgValue(root Node, node Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	return resolveRecordValues(node, "comp-call-record-arg-values", "comp-call-record-arg-value", "comp-call-record-arg-key", "comp-call-record-arg-value-type", root, invokerAncestors, currentCompCall)
}

func resolveCompCallRecordValueType(root Node, node Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	return resolveValueNode(root, node, invokerAncestors, currentCompCall)
}

func ApplyIndexes(value ResolvedValue, indexes []int) ResolvedValue {
	accessors := make([]ValueAccessor, 0, len(indexes))
	for _, index := range indexes {
		accessors = append(accessors, ValueAccessor{
			Kind:  "index",
			Index: index,
		})
	}
	return ApplyAccessors(value, accessors)
}

func ApplyAccessors(value ResolvedValue, accessors []ValueAccessor) ResolvedValue {
	result, _ := ApplyAccessorsDetailed(value, accessors)
	return result
}

func ApplyAccessorsDetailed(value ResolvedValue, accessors []ValueAccessor) (ResolvedValue, AccessError) {
	current := value
	if len(accessors) == 0 {
		return current, AccessError{}
	}

	for _, accessor := range accessors {
		if current.MissingContextKey != "" {
			return current, AccessError{}
		}

		switch accessor.Kind {
		case "key":
			if current.Type != "record" {
				return ResolvedValue{}, AccessError{
					Kind: AccessErrorInvalidKeyAccess,
				}
			}
			next, ok := current.Fields[accessor.Key]
			if !ok {
				return ResolvedValue{}, AccessError{
					Kind: AccessErrorUnknownRecordKey,
					Key:  accessor.Key,
				}
			}
			current = next
		case "index":
			if current.Type != "array" {
				return ResolvedValue{}, AccessError{
					Kind: AccessErrorInvalidIndexAccess,
				}
			}
			if accessor.Index < 0 || accessor.Index >= len(current.Items) {
				return ResolvedValue{}, AccessError{
					Kind: AccessErrorArrayIndexOutOfRange,
				}
			}
			current = current.Items[accessor.Index]
		}
	}

	return current, AccessError{}
}

func firstTypedValueNode(nodes []Node) Node {
	return FindNode(nodes, func(node Node) bool {
		return IsRuleNameOneOf(node, []string{
			"comp-context-param",
			"comp-array-param",
			"comp-record-param",
			"comp-string-param",
			"comp-integer-param",
			"comp-bool-param",
			"comp-comp-param",
			"comp-call-array-arg",
			"comp-call-context-arg",
			"comp-call-record-arg",
			"comp-call-string-arg",
			"comp-call-integer-arg",
			"comp-call-bool-arg",
			"comp-call-param-arg",
			"comp-call-comp-arg",
		})
	})
}

func getCompCallName(compCallNode Node) string {
	compCallNameNode := FindNodeByRuleName(compCallNode.Children(), "comp-call-name")
	if compCallNameNode == nil {
		return ""
	}
	return strings.TrimSpace(string(compCallNameNode.Raw()))
}
