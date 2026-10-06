package ast

import (
	"strconv"
	"strings"

	"github.com/umono-cms/compono/util"
)

type ValueAccessor struct {
	Kind  string
	Key   string
	Index int
}

func IsRuleName(node Node, name string) bool {
	return node.Rule().Name() == name
}

func IsRuleNameOneOf(node Node, names []string) bool {
	return util.InSliceString(node.Rule().Name(), names)
}

func FindNodeByRuleName(nodes []Node, name string) Node {
	return FindNode(nodes, func(node Node) bool {
		return IsRuleName(node, name)
	})
}

func FilterNodes(nodes []Node, filter func(Node) bool) []Node {
	filtered := []Node{}
	for _, node := range nodes {
		if filter(node) {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func FindNode(nodes []Node, filter func(Node) bool) Node {
	filtered := FilterNodes(nodes, filter)
	if len(filtered) > 0 {
		return filtered[0]
	}
	return nil
}

func GetAncestors(node Node) []Node {
	parent := node.Parent()
	if parent == nil {
		return []Node{}
	}
	return append([]Node{parent}, GetAncestors(parent)...)
}

func FindLocalCompDef(srcNode Node, name string) Node {
	localCompDefWrapper := FindNodeByRuleName(srcNode.Children(), "local-comp-def-wrapper")
	if localCompDefWrapper == nil {
		return nil
	}

	return FindNode(localCompDefWrapper.Children(), func(child Node) bool {
		if !IsRuleName(child, "local-comp-def") {
			return false
		}

		localCompDefHead := FindNodeByRuleName(child.Children(), "local-comp-def-head")
		if localCompDefHead == nil {
			return false
		}

		localCompName := FindNodeByRuleName(localCompDefHead.Children(), "local-comp-name")
		if localCompName == nil {
			return false
		}

		if strings.TrimSpace(string(localCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

// FindGlobalCompDef searches the scope chain starting from the enclosing globals
// of from (innermost first, then their inner scopes), stopping at isolated scopes,
// and finally checking root.
func FindGlobalCompDef(root Node, from Node, name string) Node {
	if from == nil {
		return findGlobalCompDefInWrapper(root, name)
	}

	current := from
	if !IsRuleName(current, "global-comp-def") {
		current = FindNode(GetAncestors(current), func(anc Node) bool {
			return IsRuleName(anc, "global-comp-def")
		})
	}

	for current != nil {
		if subWrapper := FindNodeByRuleName(current.Children(), "global-comp-def-wrapper"); subWrapper != nil {
			if found := findGlobalCompDefInNode(subWrapper, name); found != nil {
				return found
			}
		}

		if isIsolatedGlobal(current) {
			return nil
		}

		current = FindNode(GetAncestors(current), func(anc Node) bool {
			return IsRuleName(anc, "global-comp-def")
		})
	}

	return findGlobalCompDefInWrapper(root, name)
}

func findGlobalCompDefInWrapper(root Node, name string) Node {
	globalCompDefWrapper := FindNodeByRuleName(root.Children(), "global-comp-def-wrapper")
	if globalCompDefWrapper == nil {
		return nil
	}
	return findGlobalCompDefInNode(globalCompDefWrapper, name)
}

func findGlobalCompDefInNode(wrapper Node, name string) Node {
	return FindNode(wrapper.Children(), func(child Node) bool {
		if !IsRuleName(child, "global-comp-def") {
			return false
		}

		globalCompName := FindNodeByRuleName(child.Children(), "global-comp-name")
		if globalCompName == nil {
			return false
		}

		if strings.TrimSpace(string(globalCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

func isIsolatedGlobal(node Node) bool {
	return FindNodeByRuleName(node.Children(), "isolated-scope") != nil
}

func FindBuiltinCompDef(root Node, name string) Node {
	builtinCompDefWrapper := FindNodeByRuleName(root.Children(), "builtin-comp-wrapper")
	if builtinCompDefWrapper == nil {
		return nil
	}

	return FindNode(builtinCompDefWrapper.Children(), func(child Node) bool {
		if !IsRuleName(child, "builtin-comp") {
			return false
		}

		builtinCompName := FindNodeByRuleName(child.Children(), "builtin-comp-name")
		if builtinCompName == nil {
			return false
		}

		if strings.TrimSpace(string(builtinCompName.Raw())) != strings.TrimSpace(name) {
			return false
		}

		return true
	})
}

func FilterNodesInTree(node Node, filter func(Node) bool) []Node {
	filtered := FilterNodes(node.Children(), filter)
	for _, child := range node.Children() {
		filtered = append(filtered, FilterNodesInTree(child, filter)...)
	}
	return filtered
}

// FilterNodesInDefContent returns nodes matching filter within a definition's
// content, without entering nested sub-component definitions.
func FilterNodesInDefContent(node Node, filter func(Node) bool) []Node {
	if node == nil {
		return nil
	}

	filtered := []Node{}
	for _, child := range node.Children() {
		if IsRuleName(child, "global-comp-def-wrapper") {
			continue
		}
		if filter(child) {
			filtered = append(filtered, child)
		}
		filtered = append(filtered, FilterNodesInDefContent(child, filter)...)
	}
	return filtered
}

func GetCompParamsFromCompDef(compDef Node) []Node {
	if compDef != nil && IsRuleName(compDef, "builtin-comp") {
		compParamsNode := FindNodeByRuleName(compDef.Children(), "comp-params")
		if compParamsNode == nil {
			return []Node{}
		}
		return compParamsNode.Children()
	}

	head := GetCompDefHeadFromCompDef(compDef)
	return GetCompParamsFromCompHead(head)
}

func GetCompDefHeadFromCompDef(compDef Node) Node {
	if compDef == nil {
		return nil
	}
	return FindNode(compDef.Children(), func(node Node) bool {
		return IsRuleNameOneOf(node, []string{"local-comp-def-head", "global-comp-def-head"})
	})
}

func GetCompParamsFromCompHead(head Node) []Node {
	if head == nil {
		return []Node{}
	}
	compParamsNode := FindNodeByRuleName(head.Children(), "comp-params")
	if compParamsNode == nil {
		return []Node{}
	}
	return compParamsNode.Children()
}

func GetCompCallArgsFromCompCall(compCall Node) []Node {
	compCallArgsNode := FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgsNode == nil {
		return []Node{}
	}
	return compCallArgsNode.Children()
}

func GetCompCallArgByParamName(compCallArgs []Node, paramName string) Node {
	return FindNode(compCallArgs, func(cca Node) bool {
		return GetArgNameFromCompCallArg(cca) == paramName
	})
}

func GetTypeFromCompParam(compParam Node) string {
	compParamType := FindNodeByRuleName(compParam.Children(), "comp-param-type")
	if compParamType == nil {
		return ""
	}
	compXParam := compParamType.Children()[0]
	return strings.TrimSuffix(strings.TrimPrefix(compXParam.Rule().Name(), "comp-"), "-param")
}

func GetTypeFromCompCallArg(compCallArg Node) string {
	compCallArgType := FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	if compCallArgType == nil || len(compCallArgType.Children()) == 0 {
		return ""
	}
	compCallXArg := compCallArgType.Children()[0]
	return strings.TrimSuffix(strings.TrimPrefix(compCallXArg.Rule().Name(), "comp-call-"), "-arg")
}

func GetParamNameFromCompParam(compParam Node) string {
	compParamName := FindNodeByRuleName(compParam.Children(), "comp-param-name")
	return strings.TrimSpace(string(compParamName.Raw()))
}

// IsRequiredCompParam reports whether the given component parameter definition
// marks its parameter as required with a '!' glued to the parameter name, e.g.
// `title! = ""`. The parameter name itself never contains the '!'.
func IsRequiredCompParam(compParam Node) bool {
	if compParam == nil {
		return false
	}

	compParamName := FindNodeByRuleName(compParam.Children(), "comp-param-name")
	if compParamName == nil {
		return false
	}

	name := strings.TrimSpace(string(compParamName.Raw()))
	if name == "" {
		return false
	}

	return strings.HasPrefix(strings.TrimSpace(string(compParam.Raw())), name+"!")
}

func GetArgNameFromCompCallArg(compCallArg Node) string {
	compCallArgName := FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-name")
	return strings.TrimSpace(string(compCallArgName.Raw()))
}

func GetArgValueFromCompCallArg(compCallArg Node) string {
	compCallArgType := FindNodeByRuleName(compCallArg.Children(), "comp-call-arg-type")
	if compCallArgType == nil || len(compCallArgType.Children()) == 0 {
		return ""
	}
	compCallXArg := compCallArgType.Children()[0]
	if len(compCallXArg.Children()) == 0 {
		return ""
	}
	value := compCallXArg.Children()[0]
	rawValue := strings.TrimSpace(string(value.Raw()))
	return rawValue
}

func GetParamDefValFromCompParam(compParam Node) string {
	compParamType := FindNodeByRuleName(compParam.Children(), "comp-param-type")
	if compParamType == nil || len(compParamType.Children()) == 0 {
		return ""
	}
	compXParam := compParamType.Children()[0]
	compParamDefaValue := FindNodeByRuleName(compXParam.Children(), "comp-param-defa-value")
	if compParamDefaValue != nil {
		return strings.TrimSpace(string(compParamDefaValue.Raw()))
	}
	return strings.TrimSpace(string(compXParam.Raw()))
}

func GetParamRefName(node Node) string {
	paramRefName := FindNodeByRuleName(node.Children(), "param-ref-name")
	if paramRefName == nil {
		raw := GetParamRefRaw(node)
		name, _ := GetValuePathFromRaw(raw)
		return name
	}
	return strings.TrimSpace(string(paramRefName.Raw()))
}

func GetParamRefIndexes(node Node) []int {
	result := []int{}
	for _, accessor := range GetParamRefAccessors(node) {
		if accessor.Kind == "index" {
			result = append(result, accessor.Index)
		}
	}

	return result
}

func GetParamRefAccessors(node Node) []ValueAccessor {
	raw := GetParamRefRaw(node)
	_, accessors := GetValuePathFromRaw(raw)
	return accessors
}

func GetParamRefRaw(node Node) string {
	raw := strings.TrimSpace(string(node.Raw()))
	raw = strings.TrimPrefix(raw, "{{")
	raw = strings.TrimSuffix(raw, "}}")
	raw = strings.TrimSpace(raw)

	start := 0
	if _, ok := scanAccessorBase(raw, start); !ok {
		return ""
	}

	end := scanAccessorChain(raw, start)
	if end <= 0 {
		return ""
	}

	return strings.TrimSpace(raw[:end])
}

func GetIndexesFromRaw(raw string) []int {
	result := []int{}
	_, accessors := GetValuePathFromRaw(raw)
	for _, accessor := range accessors {
		if accessor.Kind == "index" {
			result = append(result, accessor.Index)
		}
	}
	return result
}

func GetNameFromIndexedRaw(raw string) string {
	name, _ := GetValuePathFromRaw(raw)
	return name
}

func GetValuePathFromRaw(raw string) (string, []ValueAccessor) {
	raw = strings.TrimSpace(raw)
	start := 0

	end, ok := scanAccessorBase(raw, start)
	if !ok {
		return "", nil
	}

	name := strings.TrimSpace(raw[:end])
	accessors := []ValueAccessor{}
	offset := end

	for {
		offset = skipAccessorSpaces(raw, offset)
		if offset >= len(raw) {
			break
		}

		switch raw[offset] {
		case '.':
			offset++
			offset = skipAccessorSpaces(raw, offset)
			keyEnd, ok := scanAccessorBase(raw, offset)
			if !ok {
				return name, accessors
			}
			accessors = append(accessors, ValueAccessor{
				Kind: "key",
				Key:  strings.TrimSpace(raw[offset:keyEnd]),
			})
			offset = keyEnd
		case '[':
			indexEnd := strings.Index(raw[offset:], "]")
			if indexEnd == -1 {
				return name, accessors
			}
			index, err := strconv.Atoi(strings.TrimSpace(raw[offset+1 : offset+indexEnd]))
			if err != nil {
				return name, accessors
			}
			accessors = append(accessors, ValueAccessor{
				Kind:  "index",
				Index: index,
			})
			offset = offset + indexEnd + 1
		default:
			return name, accessors
		}
	}

	return name, accessors
}

func skipAccessorSpaces(raw string, offset int) int {
	for offset < len(raw) {
		if !strings.ContainsRune(" \n\r\t", rune(raw[offset])) {
			break
		}
		offset++
	}
	return offset
}

func scanAccessorBase(raw string, offset int) (int, bool) {
	if offset >= len(raw) || raw[offset] < 'a' || raw[offset] > 'z' {
		return 0, false
	}

	offset++
	for offset < len(raw) {
		ch := raw[offset]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
			offset++
			continue
		}
		break
	}

	return offset, true
}

func scanAccessorChain(raw string, offset int) int {
	end, ok := scanAccessorBase(raw, offset)
	if !ok {
		return 0
	}

	for {
		next := skipAccessorSpaces(raw, end)
		if next >= len(raw) {
			return end
		}

		switch raw[next] {
		case '.':
			next++
			next = skipAccessorSpaces(raw, next)
			keyEnd, ok := scanAccessorBase(raw, next)
			if !ok {
				return end
			}
			end = keyEnd
		case '[':
			closeIdx := strings.Index(raw[next:], "]")
			if closeIdx == -1 {
				return end
			}
			end = next + closeIdx + 1
		default:
			return end
		}
	}
}

func FindCompDef(root Node, compCallNode Node, name string) Node {
	globalCompDefAnc := FindNode(GetAncestors(compCallNode), func(anc Node) bool {
		return IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := FindLocalCompDef(localCompDefSrc, name)
	if localCompDef != nil {
		return localCompDef
	}

	globalCompDef := FindGlobalCompDef(root, compCallNode, name)
	if globalCompDef != nil {
		return globalCompDef
	}

	builtinCompDef := FindBuiltinCompDef(root, name)
	if builtinCompDef != nil {
		return builtinCompDef
	}

	return nil
}

// FindCompDefInScope finds the component a value refers to. Local components are
// looked up in the scope the value was written in, falling back to the locals of
// the global component the node belongs to.
func FindCompDefInScope(root Node, scope Node, node Node, name string) Node {
	if scope == nil {
		scope = GetLocalCompSourceFromNode(node, root)
	}

	if localCompDef := FindLocalCompDef(scope, name); localCompDef != nil {
		return localCompDef
	}

	if currentSrc := GetLocalCompSourceFromNode(node, root); currentSrc != scope {
		if localCompDef := FindLocalCompDef(currentSrc, name); localCompDef != nil {
			return localCompDef
		}
	}

	if globalCompDef := FindGlobalCompDef(root, scope, name); globalCompDef != nil {
		return globalCompDef
	}

	return FindBuiltinCompDef(root, name)
}

func GetLocalCompSourceFromNode(node Node, root Node) Node {
	if node == nil {
		return root
	}

	if IsRuleName(node, "global-comp-def") {
		return node
	}

	globalCompDefAnc := FindNode(GetAncestors(node), func(anc Node) bool {
		return IsRuleName(anc, "global-comp-def")
	})
	if globalCompDefAnc != nil {
		return globalCompDefAnc
	}

	return root
}
