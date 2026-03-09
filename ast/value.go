package ast

import "strings"

type ResolvedValue struct {
	Type  string
	Raw   string
	Items []ResolvedValue
	Scope Node
}

func (rv ResolvedValue) IsZero() bool {
	return rv.Type == "" && rv.Raw == "" && len(rv.Items) == 0 && rv.Scope == nil
}

func ResolveCompCallArgValue(root Node, compCallArg Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	compCallArgType := FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	if compCallArgType == nil {
		return ResolvedValue{}
	}

	argTypeNode := firstTypedValueNode(compCallArgType.Children())
	if argTypeNode == nil {
		return ResolvedValue{}
	}

	if IsRuleName(argTypeNode, "comp-call-param-arg") {
		argValue := FindNodeByRuleName(argTypeNode.Children(), "comp-call-arg-value")
		if argValue == nil {
			return ResolvedValue{}
		}

		referencedParamName := GetNameFromIndexedRaw(strings.TrimSpace(string(argValue.Raw())))
		indexes := GetIndexesFromRaw(strings.TrimSpace(string(argValue.Raw())))

		remainingAncestors := invokerAncestors
		for i, anc := range invokerAncestors {
			if anc == currentCompCall {
				remainingAncestors = invokerAncestors[i+1:]
				break
			}
		}

		return ResolveParamFromAncestors(root, referencedParamName, indexes, remainingAncestors)
	}

	if IsRuleName(argTypeNode, "comp-call-array-arg") {
		resolved := resolveCompCallArrayArgValue(root, argTypeNode, invokerAncestors, currentCompCall)
		if resolved.Scope == nil {
			resolved.Scope = GetLocalCompSourceFromNode(currentCompCall, root)
		}
		return resolved
	}

	resolved := resolveLiteralValueNode(argTypeNode)
	if resolved.Scope == nil {
		resolved.Scope = GetLocalCompSourceFromNode(currentCompCall, root)
	}
	return resolved
}

func ResolveParamFromAncestors(root Node, paramName string, indexes []int, invokerAncestors []Node) ResolvedValue {
	for _, anc := range invokerAncestors {
		if !IsRuleNameOneOf(anc, []string{"block-comp-call", "inline-comp-call"}) {
			continue
		}

		compCallArgs := FindNodeByRuleName(anc.Children(), "comp-call-args")
		if compCallArgs != nil {
			compCallArg := FindNode(compCallArgs.Children(), func(cca Node) bool {
				return GetArgNameFromCompCallArg(cca) == paramName
			})

			if compCallArg != nil {
				return ApplyIndexes(ResolveCompCallArgValue(root, compCallArg, invokerAncestors, anc), indexes)
			}
		}

		if resolved := ResolveParamDefaultFromCompCall(root, anc, paramName); !resolved.IsZero() {
			return ApplyIndexes(resolved, indexes)
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

	resolved := resolveLiteralValueNode(FindNodeByRuleName(compParam.Children(), "comp-param-type"))
	if resolved.Scope == nil {
		resolved.Scope = GetLocalCompSourceFromNode(compDef, root)
	}
	return resolved
}

func resolveLiteralValueNode(node Node) ResolvedValue {
	if node == nil {
		return ResolvedValue{}
	}

	if IsRuleNameOneOf(node, []string{"comp-param-type", "comp-call-arg-type", "comp-array-param-value-type", "comp-call-array-arg-value-type"}) {
		child := firstTypedValueNode(node.Children())
		if child == nil {
			return ResolvedValue{}
		}
		return resolveLiteralValueNode(child)
	}

	switch node.Rule().Name() {
	case "comp-string-param", "comp-number-param", "comp-bool-param", "comp-comp-param":
		value := FindNodeByRuleName(node.Children(), "comp-param-defa-value")
		raw := strings.TrimSpace(string(node.Raw()))
		if value != nil {
			raw = strings.TrimSpace(string(value.Raw()))
		}
		return ResolvedValue{
			Type: strings.TrimSuffix(strings.TrimPrefix(node.Rule().Name(), "comp-"), "-param"),
			Raw:  raw,
		}
	case "comp-call-string-arg", "comp-call-number-arg", "comp-call-bool-arg", "comp-call-comp-arg", "comp-call-param-arg":
		value := FindNodeByRuleName(node.Children(), "comp-call-arg-value")
		raw := strings.TrimSpace(string(node.Raw()))
		if value != nil {
			raw = strings.TrimSpace(string(value.Raw()))
		}
		return ResolvedValue{
			Type: strings.TrimSuffix(strings.TrimPrefix(node.Rule().Name(), "comp-call-"), "-arg"),
			Raw:  raw,
		}
	case "comp-array-param":
		return resolveArrayValues(node, "comp-array-param-values", "comp-array-param-value", "comp-array-param-value-type")
	case "comp-call-array-arg":
		return resolveArrayValues(node, "comp-call-array-arg-values", "comp-call-array-arg-value", "comp-call-array-arg-value-type")
	default:
		if len(node.Children()) == 1 {
			return resolveLiteralValueNode(node.Children()[0])
		}
	}

	return ResolvedValue{}
}

func resolveArrayValues(node Node, valuesRuleName, valueRuleName, valueTypeRuleName string) ResolvedValue {
	values := FindNodeByRuleName(node.Children(), valuesRuleName)
	if values == nil {
		return ResolvedValue{Type: "array", Items: []ResolvedValue{}}
	}

	items := []ResolvedValue{}
	for _, valueNode := range values.Children() {
		if !IsRuleName(valueNode, valueRuleName) {
			continue
		}
		resolved := resolveLiteralValueNode(FindNodeByRuleName(valueNode.Children(), valueTypeRuleName))
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

func resolveCompCallArrayArgValue(root Node, node Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	values := FindNodeByRuleName(node.Children(), "comp-call-array-arg-values")
	if values == nil {
		return ResolvedValue{Type: "array", Items: []ResolvedValue{}}
	}

	items := []ResolvedValue{}
	for _, valueNode := range values.Children() {
		if !IsRuleName(valueNode, "comp-call-array-arg-value") {
			continue
		}

		valueTypeNode := FindNodeByRuleName(valueNode.Children(), "comp-call-array-arg-value-type")
		if valueTypeNode == nil {
			continue
		}

		item := resolveCompCallArrayValueType(root, valueTypeNode, invokerAncestors, currentCompCall)
		if item.IsZero() {
			continue
		}
		if item.Scope == nil {
			item.Scope = GetLocalCompSourceFromNode(currentCompCall, root)
		}
		items = append(items, item)
	}

	return ResolvedValue{
		Type:  "array",
		Items: items,
	}
}

func resolveCompCallArrayValueType(root Node, node Node, invokerAncestors []Node, currentCompCall Node) ResolvedValue {
	valueNode := firstTypedValueNode(node.Children())
	if valueNode == nil {
		return ResolvedValue{}
	}

	if IsRuleName(valueNode, "comp-call-param-arg") {
		argValue := FindNodeByRuleName(valueNode.Children(), "comp-call-arg-value")
		if argValue == nil {
			return ResolvedValue{}
		}

		raw := strings.TrimSpace(string(argValue.Raw()))
		referencedParamName := GetNameFromIndexedRaw(raw)
		indexes := GetIndexesFromRaw(raw)

		remainingAncestors := invokerAncestors
		for i, anc := range invokerAncestors {
			if anc == currentCompCall {
				remainingAncestors = invokerAncestors[i+1:]
				break
			}
		}

		return ResolveParamFromAncestors(root, referencedParamName, indexes, remainingAncestors)
	}

	if IsRuleName(valueNode, "comp-call-array-arg") {
		return resolveCompCallArrayArgValue(root, valueNode, invokerAncestors, currentCompCall)
	}

	return resolveLiteralValueNode(valueNode)
}

func ApplyIndexes(value ResolvedValue, indexes []int) ResolvedValue {
	current := value
	if len(indexes) == 0 {
		return current
	}

	for _, index := range indexes {
		if current.Type != "array" || index < 0 || index >= len(current.Items) {
			return ResolvedValue{}
		}
		current = current.Items[index]
	}

	return current
}

func firstTypedValueNode(nodes []Node) Node {
	return FindNode(nodes, func(node Node) bool {
		return IsRuleNameOneOf(node, []string{
			"comp-array-param",
			"comp-string-param",
			"comp-number-param",
			"comp-bool-param",
			"comp-comp-param",
			"comp-call-array-arg",
			"comp-call-string-arg",
			"comp-call-number-arg",
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
