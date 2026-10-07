package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/builtin"
	"github.com/umono-cms/compono/util"
)

// maxBindingWalkDepth bounds the static walk over component frames. Recursive
// components are reported by the infinite loop analyzers.
const maxBindingWalkDepth = 32

// boundCompValue is a component value with bound arguments, e.g. X(a = 1).
type boundCompValue struct {
	name    string
	compDef ast.Node
	args    []ast.Node
}

// boundArgumentAnalyzer reports the errors of bound arguments that can only be
// found by following component values to where they are rendered: duplicate
// arguments and invalid arguments of bound built-ins.
type boundArgumentAnalyzer struct{}

func (boundArgumentAnalyzer) Diagnose(ctx *wrapContext, node ast.Node) (diagnostic, bool) {
	if !ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call", "param-ref"}) {
		return diagnostic{}, false
	}

	diag, ok := getBindingIssues(ctx)[node]
	if !ok {
		return diagnostic{}, false
	}

	if ast.IsRuleName(node, "param-ref") {
		diag.block = blockForParamRef(ctx, node)
	} else {
		diag.block = blockFromRuleName(ctx, node)
	}
	return diag, true
}

// getBindingIssues walks every component call of the root content like the
// renderer does and collects the errors of bound arguments. An error is placed
// on the call the bound component value is written in. A value bound in a
// default value is placed on the call that used the default value.
func getBindingIssues(ctx *wrapContext) map[ast.Node]diagnostic {
	if ctx.bindingIssues != nil {
		return ctx.bindingIssues
	}
	ctx.bindingIssues = map[ast.Node]diagnostic{}

	rootContent := ast.FindNodeByRuleName(ctx.root.Children(), "root-content")
	if rootContent == nil {
		return ctx.bindingIssues
	}

	for _, compCall := range ast.FilterNodesInTree(rootContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	}) {
		walkBindingFrames(ctx, []ast.Node{compCall})
	}

	return ctx.bindingIssues
}

func walkBindingFrames(ctx *wrapContext, chain []ast.Node) {
	if len(chain) > maxBindingWalkDepth {
		return
	}

	frame := chain[0]
	after := chain[1:]

	compDef := ast.FindFrameCompDef(ctx.root, frame, after)
	if compDef == nil {
		return
	}

	if ast.IsRuleName(compDef, "builtin-comp") {
		checkBoundBuiltinFrame(ctx, frame, after, compDef)
		return
	}

	content := getCompDefContent(compDef)
	if content == nil {
		return
	}

	for _, next := range ast.FilterNodesInTree(content, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call", "param-ref"})
	}) {
		if nodeInChain(next, chain) {
			continue
		}

		walkBindingFrames(ctx, append([]ast.Node{next}, chain...))
	}
}

// checkBoundBuiltinFrame validates a bound built-in. Bound arguments and the
// arguments given by the caller are validated together.
func checkBoundBuiltinFrame(ctx *wrapContext, frame ast.Node, after []ast.Node, compDef ast.Node) {
	if !ast.IsRuleName(frame, "param-ref") {
		return
	}

	value := ast.ResolveFrameCompValue(ctx.root, frame, after)
	if value.Bound == nil {
		return
	}

	definition, ok := builtin.FindDefinition(value.Raw)
	if !ok {
		return
	}

	for _, param := range definition.Params {
		resolved := ast.ResolveFrameParam(ctx.root, frame, param.Name, after)
		if resolved.IsZero() || resolvedValueMissingContextKey(resolved) != "" || builtin.MatchesResolvedValue(param.Schema, resolved) {
			continue
		}

		diag := diagnostic{
			title:   "Invalid built-in arguments",
			message: "The parameter **" + param.Name + "** does not match the schema of the built-in component **" + value.Raw + "**.",
		}
		if paramDiagnostic := builtinParamDiagnostic(param, resolved); paramDiagnostic.Title != "" {
			diag.title = paramDiagnostic.Title
			diag.message = paramDiagnostic.Message
		}

		addBindingIssue(ctx, value, diag)
		return
	}
}

func addBindingIssue(ctx *wrapContext, value ast.ResolvedValue, diag diagnostic) {
	target := value.Bound.WrittenIn()
	if target == nil {
		return
	}
	if _, exists := ctx.bindingIssues[target]; exists {
		return
	}
	ctx.bindingIssues[target] = diag
}

func nodeInChain(node ast.Node, chain []ast.Node) bool {
	for _, n := range chain {
		if n == node {
			return true
		}
	}
	return false
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
