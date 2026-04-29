package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

func unknownCompCall() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isUnknownComponent(),
		},
		title:   staticTitle("Unknown component"),
		message: unknownCompCallMsg,
		block:   blockFromRuleName,
	}
}

func unknownCompParamCall() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasUnknownResolvedCompArg(),
		},
		title:   staticTitle("Unknown component"),
		message: unknownCompParamCallMsg,
		block:   blockFromRuleName,
	}
}

func blockCompInsideInline() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("inline-comp-call"),
			isKnownComponent(),
			callsBlockComponent(),
		},
		title:   staticTitle("Invalid component usage"),
		message: blockCompInsideInlineMsg,
		block:   neverBlock,
	}
}

func blockParamCompInsideInline() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			resolvedCompArgUsedAsInlineButIsBlock(),
		},
		title:   staticTitle("Invalid component usage"),
		message: blockParamCompInsideInlineMsg,
		block:   blockFromRuleName,
	}
}

func isLegacyNotCompStandalone() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		if getParamRefNameStr(node) != "param" {
			return false
		}
		if !blockForParamRef(nil, node) {
			return false
		}
		compDef := findEnclosingCompDef(node)
		if compDef == nil || !ast.IsRuleName(compDef, "global-comp-def") {
			return false
		}

		for _, info := range getCompDefParamInfos(compDef) {
			if info.name == "param" && info.typ == "string" && info.defVal == "I am a string parameter" {
				return true
			}
		}
		return false
	}
}

func isCompParamRefNode() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		compDef := findEnclosingCompDef(node)
		if compDef == nil {
			return false
		}
		return isCompParamRefInCompDef(compDef, node)
	}
}

func isCompParamRefInCompDef(compDef ast.Node, node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}

	paramName := getParamRefNameStr(node)
	if paramName == "" {
		return false
	}

	for _, info := range getCompDefParamInfos(compDef) {
		if info.name != paramName {
			continue
		}
		return info.typ == "comp" || info.typ == ""
	}

	return false
}

func isUnknownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		return findCompDef(ctx.root, node, compCallName) == nil
	}
}

func isKnownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !isUnknownComponent()(ctx, node)
	}
}

func hasUnknownResolvedCompArg() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		compDef := findCompDef(ctx.root, compCall, compCallName)
		if compDef == nil {
			return false
		}
		return len(getUnknownResolvedCompArgs(ctx, compCall, compDef)) > 0
	}
}

func getUnknownResolvedCompArgs(ctx *wrapContext, compCall ast.Node, compDef ast.Node) []string {
	if compDef == nil {
		return nil
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return nil
	}

	usedCompParamNames := map[string]struct{}{}
	for _, paramCompCall := range ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
		return isCompParamRefInCompDef(compDef, node)
	}) {
		name := getParamCompCallNameStr(paramCompCall)
		if name == "" {
			continue
		}
		usedCompParamNames[name] = struct{}{}
	}

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	if len(resolvedCompArgs) == 0 {
		return nil
	}

	explicitCompArgs := getExplicitCompArgMap(compCall)
	var unknowns []string

	for _, info := range getCompDefParamInfos(compDef) {
		if info.typ != "comp" {
			continue
		}
		if _, used := usedCompParamNames[info.name]; !used {
			continue
		}

		value := resolvedCompArgs[info.name]
		if value == "" || strings.HasPrefix(value, "$") {
			continue
		}

		lookupScope := compCall
		if _, ok := explicitCompArgs[info.name]; !ok && ast.IsRuleName(compDef, "global-comp-def") {
			if globalCompContent := getCompDefContent(compDef); globalCompContent != nil {
				lookupScope = globalCompContent
			}
		}

		if findCompDef(ctx.root, lookupScope, value) == nil && !util.InSliceString(value, unknowns) {
			unknowns = append(unknowns, value)
		}
	}

	return unknowns
}

func callsBlockComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		if compCallName == "" {
			return false
		}

		compDef := findCompDef(ctx.root, node, compCallName)
		if compDef == nil {
			return false
		}

		return isBlockComponent(compDef)
	}
}

func resolvedCompArgUsedAsInlineButIsBlock() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		return getResolvedInlineBlockCompName(ctx, compCall) != ""
	}
}

func getResolvedInlineBlockCompName(ctx *wrapContext, compCall ast.Node) string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return ""
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return ""
	}

	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return ""
	}

	resolved := resolveCompArgValues(ctx, compCall)
	explicitCompArgs := getExplicitCompArgMap(compCall)
	explicitParamArgs := getExplicitParamArgMap(compCall)

	inlineParamCalls := ast.FilterNodesInTree(compDefContent, func(n ast.Node) bool {
		return ast.IsRuleName(n, "param-ref") && isInlineParamRefNode(n)
	})

	for _, ipc := range inlineParamCalls {
		hasKeyAccessor := false
		for _, accessor := range ast.GetParamRefAccessors(ipc) {
			if accessor.Kind == "key" {
				hasKeyAccessor = true
				break
			}
		}
		if hasKeyAccessor {
			continue
		}

		ipcName := getParamRefNameStr(ipc)
		if ipcName == "" {
			continue
		}

		resolvedValue, ok := resolveInlineParamRefValue(ctx, compCall, ipc, resolved)
		if !ok || resolvedValue.Type != "comp" || resolvedValue.Raw == "" {
			continue
		}

		resolvedCompName := resolvedValue.Raw

		lookupScope := compCall
		_, hasExplicitCompArg := explicitCompArgs[ipcName]
		_, hasExplicitParamArg := explicitParamArgs[ipcName]
		if !hasExplicitCompArg && !hasExplicitParamArg && ast.IsRuleName(compDef, "global-comp-def") {
			if globalCompContent := getCompDefContent(compDef); globalCompContent != nil {
				lookupScope = globalCompContent
			}
		}

		argCompDef := findCompDef(ctx.root, lookupScope, resolvedCompName)
		if argCompDef == nil {
			continue
		}

		if isBlockComponent(argCompDef) {
			return resolvedCompName
		}

		if !hasExplicitCompArg || hasExplicitParamArg || !ast.IsRuleName(compDef, "global-comp-def") {
			continue
		}

		globalCompContent := getCompDefContent(compDef)
		if globalCompContent == nil || len(globalCompContent.Children()) <= 1 {
			continue
		}

		globalScopedCompDef := findCompDef(ctx.root, globalCompContent, resolvedCompName)
		if globalScopedCompDef == nil {
			return resolvedCompName
		}
	}

	return ""
}

func resolveInlineParamRefValue(
	ctx *wrapContext,
	compCall ast.Node,
	paramRef ast.Node,
	resolved map[string]string,
) (ast.ResolvedValue, bool) {
	paramName := getParamRefNameStr(paramRef)
	if paramName == "" {
		return ast.ResolvedValue{}, false
	}

	accessors := ast.GetParamRefAccessors(paramRef)
	if len(accessors) > 0 {
		if explicitArg := ast.GetCompCallArgByParamName(ast.GetCompCallArgsFromCompCall(compCall), paramName); explicitArg != nil {
			return ast.ApplyAccessors(
				ast.ResolveCompCallArgValue(ctx.root, explicitArg, ast.GetAncestors(compCall), compCall),
				accessors,
			), true
		}

		return ast.ApplyAccessors(
			ast.ResolveParamDefaultFromCompCall(ctx.root, compCall, paramName),
			accessors,
		), true
	}

	resolvedCompName, ok := resolved[paramName]
	if !ok || resolvedCompName == "" || strings.HasPrefix(resolvedCompName, "$") {
		return ast.ResolvedValue{}, false
	}

	return ast.ResolvedValue{
		Type:  "comp",
		Raw:   resolvedCompName,
		Scope: ast.GetLocalCompSourceFromNode(compCall, ctx.root),
	}, true
}

func isNotCompParamCompCall() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		if len(ast.GetParamRefAccessors(node)) > 0 {
			return false
		}

		paramName := getParamCompCallNameStr(node)
		if paramName == "" {
			return false
		}

		compDef := findEnclosingCompDef(node)
		if compDef == nil {
			return false
		}

		for _, info := range getCompDefParamInfos(compDef) {
			if info.name != paramName {
				continue
			}
			return info.typ != "" && info.typ != "comp"
		}

		return false
	}
}
