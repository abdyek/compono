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
