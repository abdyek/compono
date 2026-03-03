package ast

import (
	"strings"

	"github.com/umono-cms/compono/util"
)

func IsRuleName(node Node, name string) bool {
	if node.Rule().Name() != name {
		return false
	}
	return true
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

func FindGlobalCompDef(root Node, name string) Node {
	globalCompDefWrapper := FindNodeByRuleName(root.Children(), "global-comp-def-wrapper")
	if globalCompDefWrapper == nil {
		return nil
	}

	return FindNode(globalCompDefWrapper.Children(), func(child Node) bool {
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

func FilterNodesInTree(node Node, filter func(Node) bool) []Node {
	filtered := FilterNodes(node.Children(), filter)
	for _, child := range node.Children() {
		filtered = append(filtered, FilterNodesInTree(child, filter)...)
	}
	return filtered
}

func GetCompParamsFromCompDef(compDef Node) []Node {
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
	compXParam := compParamType.Children()[0]
	compParamDefaValue := compXParam.Children()[0]
	return strings.TrimSpace(string(compParamDefaValue.Raw()))
}
