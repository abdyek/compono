package errwrap

import (
	"sort"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

func isCalledByItself() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		if compCallName == "" {
			return false
		}

		calledCompDef := findCompDef(ctx.root, node, compCallName)
		if calledCompDef == nil {
			return false
		}

		for _, anc := range ast.GetAncestors(node) {
			if anc == calledCompDef {
				return true
			}
		}

		return false
	}
}

func isCalledByChain() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		for _, chain := range ctx.compCallChains {
			for j := 1; j < len(chain); j++ {
				repeated := false
				for i := 0; i < j; i++ {
					if chain[i] == chain[j] {
						repeated = true
						break
					}
				}

				if repeated && chain[j-1] == compCall {
					if !compCallHasCycle(ctx, chain[j]) &&
						getCompCallNameStr(chain[j]) != getCompCallNameStr(compCall) {
						continue
					}
					return true
				}
			}
		}

		return false
	}
}

func compCallHasCycle(ctx *wrapContext, compCall ast.Node) bool {
	if ctx.compCallCycleCache == nil {
		ctx.compCallCycleCache = map[ast.Node]bool{}
	}
	if cached, ok := ctx.compCallCycleCache[compCall]; ok {
		return cached
	}

	startName := getCompCallNameStr(compCall)
	if startName == "" {
		ctx.compCallCycleCache[compCall] = false
		return false
	}

	if findCompDef(ctx.root, compCall, startName) == nil {
		ctx.compCallCycleCache[compCall] = false
		return false
	}

	var dfs func(callNode ast.Node, path []string) bool
	dfs = func(callNode ast.Node, path []string) bool {
		callName := getCompCallNameStr(callNode)
		if callName == "" {
			return false
		}

		def := findCompDef(ctx.root, callNode, callName)
		if def == nil {
			return false
		}

		content := getCompDefContent(def)
		if content == nil {
			return false
		}

		children := ast.FilterNodesInTree(content, func(node ast.Node) bool {
			return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
		})

		nextPath := append(append([]string{}, path...), callName)
		for _, child := range children {
			childName := getCompCallNameStr(child)
			if childName == "" {
				continue
			}

			if util.InSliceString(childName, nextPath) {
				return true
			}

			if dfs(child, nextPath) {
				return true
			}
		}

		return false
	}

	hasCycle := dfs(compCall, nil)
	ctx.compCallCycleCache[compCall] = hasCycle
	return hasCycle
}

func takesItselfAsArgOrDefault() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		if compCallName == "" {
			return false
		}

		compDef := findCompDef(ctx.root, compCall, compCallName)
		if compDef == nil {
			return false
		}

		compDefContent := getCompDefContent(compDef)
		if compDefContent == nil {
			return false
		}

		resolved := resolveCompArgValues(ctx, compCall)

		for paramName, resolvedValue := range resolved {
			if strings.HasPrefix(resolvedValue, "$") {
				continue
			}

			if resolvedValue != compCallName {
				continue
			}

			paramCompCalls := ast.FilterNodesInTree(compDefContent, func(n ast.Node) bool {
				if !isCompParamRefInCompDef(compDef, n) {
					return false
				}
				return getParamCompCallNameStr(n) == paramName
			})

			if len(paramCompCalls) > 0 {
				return true
			}
		}

		return false
	}
}

func isClosingParamCompCallInCycle() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return getClosingParamCompCallTargetName(ctx, node) != ""
	}
}

func getClosingParamCompCallTargetName(ctx *wrapContext, node ast.Node) string {
	return getParamCycleClosers(ctx)[node]
}

func getParamCycleClosers(ctx *wrapContext) map[ast.Node]string {
	if ctx.paramCycleClosers != nil {
		return ctx.paramCycleClosers
	}

	ctx.paramCycleClosers = map[ast.Node]string{}

	rootContent := ast.FindNodeByRuleName(ctx.root.Children(), "root-content")
	if rootContent == nil {
		return ctx.paramCycleClosers
	}

	rootCompCalls := ast.FilterNodesInTree(rootContent, func(node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"})
	})

	for _, rootCompCall := range rootCompCalls {
		rootCompName := getCompCallNameStr(rootCompCall)
		if rootCompName == "" {
			continue
		}

		path := map[string]bool{}
		var dfs func(callNode ast.Node, compName string, resolved map[string]string)

		dfs = func(callNode ast.Node, compName string, resolved map[string]string) {
			signature := makeResolvedCallSignature(compName, resolved)
			if path[signature] {
				return
			}

			path[signature] = true
			defer delete(path, signature)

			compDef := findCompDef(ctx.root, callNode, compName)
			if compDef == nil {
				return
			}

			compDefContent := getCompDefContent(compDef)
			if compDefContent == nil {
				return
			}

			paramCompCalls := ast.FilterNodesInTree(compDefContent, func(node ast.Node) bool {
				return isCompParamRefInCompDef(compDef, node)
			})

			for _, paramCompCall := range paramCompCalls {
				paramName := getParamCompCallNameStr(paramCompCall)
				if paramName == "" {
					continue
				}

				targetCompName := resolved[paramName]
				if targetCompName == "" || strings.HasPrefix(targetCompName, "$") {
					continue
				}

				if findCompDef(ctx.root, paramCompCall, targetCompName) == nil {
					continue
				}

				nextResolved := resolveCompArgValuesForCallTarget(ctx, paramCompCall, targetCompName, resolved)
				nextSignature := makeResolvedCallSignature(targetCompName, nextResolved)
				if path[nextSignature] {
					if _, exists := ctx.paramCycleClosers[paramCompCall]; !exists {
						ctx.paramCycleClosers[paramCompCall] = targetCompName
					}
					continue
				}

				dfs(paramCompCall, targetCompName, nextResolved)
			}
		}

		dfs(rootCompCall, rootCompName, resolveCompArgValues(ctx, rootCompCall))
	}

	return ctx.paramCycleClosers
}

func resolveCompArgValuesForCallTarget(
	ctx *wrapContext,
	callNode ast.Node,
	targetCompName string,
	parentResolvedCompArgs map[string]string,
) map[string]string {
	targetCompDef := findCompDef(ctx.root, callNode, targetCompName)
	if targetCompDef == nil {
		return nil
	}

	resolved := map[string]string{}
	for paramName, defaultValue := range getCompDefCompParamDefaults(targetCompDef) {
		resolved[paramName] = defaultValue
	}

	for _, arg := range ast.GetCompCallArgsFromCompCall(callNode) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}

		argName := ast.GetArgNameFromCompCallArg(arg)
		if argName == "" {
			continue
		}

		argType := ast.GetTypeFromCompCallArg(arg)
		switch argType {
		case "comp":
			resolved[argName] = ast.GetArgValueFromCompCallArg(arg)
		case "param":
			argValue := ast.GetArgValueFromCompCallArg(arg)
			if argValue == "" {
				continue
			}

			if parentResolvedCompArgs != nil {
				if forwardedValue, ok := parentResolvedCompArgs[argValue]; ok {
					resolved[argName] = forwardedValue
					continue
				}
			}

			resolved[argName] = "$" + argValue
		}
	}

	if len(resolved) == 0 {
		return nil
	}

	return resolved
}

func makeResolvedCallSignature(compName string, resolved map[string]string) string {
	if compName == "" {
		return ""
	}

	if len(resolved) == 0 {
		return compName
	}

	keys := make([]string, 0, len(resolved))
	for k := range resolved {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString(compName)
	for _, k := range keys {
		b.WriteString("|")
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(resolved[k])
	}

	return b.String()
}
