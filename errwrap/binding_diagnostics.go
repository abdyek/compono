package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

// boundCompValue is a component value with bound arguments, e.g. X(a = 1).
type boundCompValue struct {
	name    string
	compDef ast.Node
	args    []ast.Node
}

// getBoundCompValues returns the component values with bound arguments written
// under the node, including nested ones.
func getBoundCompValues(root ast.Node, node ast.Node) []boundCompValue {
	result := []boundCompValue{}
	for _, boundArgs := range ast.FilterNodesInTree(node, func(n ast.Node) bool {
		return ast.IsRuleName(n, "comp-bound-args")
	}) {
		compValue := boundArgs.Parent()
		if compValue == nil {
			continue
		}

		nameNode := ast.FindNode(compValue.Children(), func(child ast.Node) bool {
			return ast.IsRuleNameOneOf(child, []string{"comp-call-arg-value", "comp-param-defa-value"})
		})
		if nameNode == nil {
			continue
		}
		name := strings.TrimSpace(string(nameNode.Raw()))

		result = append(result, boundCompValue{
			name:    name,
			compDef: ast.FindCompDef(root, compValue, name),
			args: ast.FilterNodes(boundArgs.Children(), func(child ast.Node) bool {
				return ast.IsRuleName(child, "comp-call-arg")
			}),
		})
	}
	return result
}

// getBindingSources returns the nodes whose bound arguments are given to the
// call: its arguments and the default values it uses.
func getBindingSources(root ast.Node, call ast.Node) []ast.Node {
	sources := ast.GetCompCallArgsFromCompCall(call)
	if !ast.IsRuleNameOneOf(call, []string{"block-comp-call", "inline-comp-call"}) {
		return sources
	}

	compDef := ast.FindCompDef(root, call, getCompCallNameStr(call))
	if compDef == nil || ast.IsRuleName(compDef, "builtin-comp") {
		return sources
	}

	for _, compParam := range ast.GetCompParamsFromCompDef(compDef) {
		if ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(call), ast.GetParamNameFromCompParam(compParam)) != nil {
			continue
		}
		sources = append(sources, compParam)
	}
	return sources
}

func getUndefinedBoundArgNames(root ast.Node, call ast.Node) []string {
	names := []string{}
	for _, source := range getBindingSources(root, call) {
		for _, value := range getBoundCompValues(root, source) {
			if value.compDef == nil {
				continue
			}

			definedParams := getCompDefParamNames(value.compDef)
			for _, arg := range value.args {
				argName := ast.GetArgNameFromCompCallArg(arg)
				if util.InSliceString(argName, definedParams) {
					continue
				}
				names = appendUniqueStrings(names, argName)
			}
		}
	}

	return appendUniqueStrings(names, getUndefinedParamRefNamesInDefaultBindings(root, call)...)
}

// getUndefinedParamRefNamesInDefaultBindings returns the parameters referenced by
// the bound arguments of the default values a call uses, that are not defined
// by the component of those default values.
func getUndefinedParamRefNamesInDefaultBindings(root ast.Node, call ast.Node) []string {
	names := []string{}
	for _, source := range getBindingSources(root, call) {
		if !ast.IsRuleName(source, "comp-param") {
			continue
		}

		compDef := findEnclosingCompDef(source)
		if compDef == nil {
			continue
		}

		definedParams := getCompDefParamNames(compDef)
		for _, paramArg := range ast.FilterNodesInTree(source, func(node ast.Node) bool {
			return ast.IsRuleName(node, "comp-call-param-arg")
		}) {
			refName, _ := ast.GetValuePathFromRaw(string(paramArg.Raw()))
			if refName == "" || util.InSliceString(refName, definedParams) {
				continue
			}
			names = appendUniqueStrings(names, refName)
		}
	}
	return names
}

func getWrongTypeBoundArgNames(root ast.Node, call ast.Node) []string {
	names := []string{}
	for _, source := range getBindingSources(root, call) {
		for _, value := range getBoundCompValues(root, source) {
			if value.compDef == nil {
				continue
			}

			paramTypeMap := getCompDefParamTypeMap(value.compDef)
			for _, arg := range value.args {
				argName := ast.GetArgNameFromCompCallArg(arg)
				expectedType, ok := paramTypeMap[argName]
				if !ok || expectedType == "" {
					continue
				}

				actualType := ast.GetTypeFromCompCallArg(arg)
				if actualType == "context" {
					actualType = ast.ResolveCompCallArgValue(root, arg, nil, arg).Type
				}
				if actualType == "" || actualType == "param" || actualType == expectedType {
					continue
				}

				names = appendUniqueStrings(names, argName)
			}
		}
	}
	return names
}

func duplicateArgumentMsg(name string, compName string) string {
	return "The parameter **" + name + "** of component **" + compName + "** is already bound."
}
