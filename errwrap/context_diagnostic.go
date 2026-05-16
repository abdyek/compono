package errwrap

import "github.com/umono-cms/compono/ast"

type contextRefAnalyzer struct{}

func (contextRefAnalyzer) Diagnose(ctx *wrapContext, node ast.Node) (diagnostic, bool) {
	if !isDirectContextRefNode(node) {
		return diagnostic{}, false
	}

	resolved, accessErr := resolveContextRefValueDetailed(ctx.root, node)
	if resolved.MissingContextKey != "" {
		return diagnostic{
			title:   "Unknown key",
			message: unknownContextKeyMsg(resolved.MissingContextKey),
			block:   false,
		}, true
	}

	if accessErr.Kind == "" && len(ast.GetContextAccessors(node)) == 0 && (resolved.Type == "array" || resolved.Type == "record") {
		return diagnostic{
			title:   "Invalid parameter usage",
			message: contextRefInvalidUsageMessage(resolved),
			block:   false,
		}, true
	}

	if accessErr.Kind == ast.AccessErrorUnknownRecordKey {
		return diagnostic{
			title:   "Unknown record key",
			message: unknownRecordKeyMsg(accessErr.Key),
			block:   false,
		}, true
	}

	if accessErr.Kind == ast.AccessErrorArrayIndexOutOfRange {
		return diagnostic{
			title:   "Array index out of range",
			message: contextArrayIndexOutOfRangeMsg(),
			block:   false,
		}, true
	}

	if accessErr.Kind == ast.AccessErrorInvalidIndexAccess {
		return diagnostic{
			title:   "Invalid context access",
			message: invalidContextIndexAccessMsg(ast.GetContextKey(node)),
			block:   false,
		}, true
	}

	if accessErr.Kind == ast.AccessErrorInvalidKeyAccess {
		return diagnostic{
			title:   "Invalid context access",
			message: invalidContextKeyAccessMsg(ast.GetContextKey(node)),
			block:   false,
		}, true
	}

	return diagnostic{}, false
}

func isDirectContextRefNode(node ast.Node) bool {
	if !ast.IsRuleName(node, "context-ref") {
		return false
	}

	return ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{
			"comp-param-type",
			"comp-call-arg-type",
			"comp-array-param-value-type",
			"comp-call-array-arg-value-type",
			"comp-record-param-value-type",
			"comp-call-record-arg-value-type",
			"context-value-type",
		})
	}) == nil
}

func contextRefInvalidUsageMessage(resolved ast.ResolvedValue) string {
	if resolved.Type == "array" {
		return "The context value is an array and cannot be rendered directly."
	}
	return "The context value is a record and cannot be rendered directly."
}

func resolveContextRefValueDetailed(root ast.Node, node ast.Node) (ast.ResolvedValue, ast.AccessError) {
	key := ast.GetContextKey(node)
	if key == "" {
		return ast.ResolvedValue{}, ast.AccessError{}
	}

	resolved := ast.ResolveContextValue(root, key)
	if resolved.MissingContextKey != "" {
		return resolved, ast.AccessError{}
	}

	return ast.ApplyAccessorsDetailed(resolved, ast.GetContextAccessors(node))
}
