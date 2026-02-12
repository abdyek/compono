package errwrap

import (
	"regexp"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

// TODO: Remove this const in the codes. It is not clean code
var (
	globalCompParamLineRE = regexp.MustCompile(`^([a-z][a-z0-9-]*)\s*=\s*(".*?"|\d+(?:\.\d+)?|true|false|[A-Z0-9]+(?:_[A-Z0-9]+)*)\s*$`)
	numberValueRE         = regexp.MustCompile(`^\d+(?:\.\d+)?$`)
	compValueRE           = regexp.MustCompile(`^[A-Z0-9]+(?:_[A-Z0-9]+)*$`)
)

type compParamInfo struct {
	name   string
	typ    string
	defVal string
}

type wrapContext struct {
	root           ast.Node
	compCallChains [][]ast.Node
}

type wrapRule struct {
	conditions []func(ctx *wrapContext, node ast.Node) bool
	title      func(ctx *wrapContext, node ast.Node) string
	message    func(ctx *wrapContext, node ast.Node) string
	block      func(ctx *wrapContext, node ast.Node) bool
}

func wrapRules() []wrapRule {
	return []wrapRule{
		infiniteBlockCompCallByItself(),
		infiniteInlineCompCallByItself(),
		infiniteCompCallByChain(),
		infiniteCompCallByParam(),
		unknownCompCall(),
		unknownCompParamCall(),
		blockCompInsideInline(),
		blockParamCompInsideInline(),
		undefinedCompCallArg(),
		wrongArgType(),
		paramRefInRootContent(),
		paramCompCallInRootContent(),
		undefinedParamRef(),
		notCompParamCompCall(),
		undefinedParamCompCall(),
	}
}

func infiniteBlockCompCallByItself() wrapRule {
	return wrapRule{
		conditions: []func(_ *wrapContext, node ast.Node) bool{
			isRuleName("block-comp-call"),
			isCalledByItself(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   alwaysBlock,
	}
}

func infiniteInlineCompCallByItself() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("inline-comp-call"),
			isCalledByItself(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallMsg,
		block:   neverBlock,
	}
}

func infiniteCompCallByChain() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			isCalledByChain(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteChainCompCallMsg,
		block:   blockFromRuleName,
	}
}

func infiniteCompCallByParam() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			takesItselfAsArgOrDefault(),
		},
		title:   staticTitle("Infinite component call"),
		message: infiniteCompCallByParamMsg,
		block:   blockFromRuleName,
	}
}

func unknownCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isUnknownComponent(),
		},
		title:   staticTitle("Unknown component"),
		message: unknownCompCallMsg,
		block:   blockFromRuleName,
	}
}

func unknownCompParamCall() wrapRule {
	return wrapRule{
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

func blockCompInsideInline() wrapRule {
	return wrapRule{
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

func blockParamCompInsideInline() wrapRule {
	return wrapRule{
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

// TODO: Complete it, message func creates wrong info
func undefinedCompCallArg() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasUndefinedArgs(),
		},
		title:   staticTitle("Undefined parameter"),
		message: undefinedArgMsgFn,
		block:   blockFromRuleName,
	}
}

func wrongArgType() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasWrongTypeArgs(),
		},
		title:   staticTitle("Wrong argument type"),
		message: wrongArgTypeMsgFn,
		block:   blockFromRuleName,
	}
}

func paramRefInRootContent() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			isInsideRootContent(),
		},
		title:   staticTitle("Invalid parameter usage"),
		message: paramRefInRootMsg,
		block:   neverBlock,
	}
}

func paramCompCallInRootContent() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-param-comp-call", "inline-param-comp-call"),
			isInsideRootContent(),
		},
		title:   staticTitle("Parameter component call in root content"),
		message: paramCompCallInRootMsg,
		block:   blockFromRuleName,
	}
}

func undefinedParamRef() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleName("param-ref"),
			not(isInsideRootContent()),
			isUndefinedParamRef(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamRefMsg,
		block:   blockFromRuleName,
	}
}

func undefinedParamCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-param-comp-call", "inline-param-comp-call"),
			not(isInsideRootContent()),
			isUndefinedParamCompCall(),
		},
		title:   staticTitle("Unknown parameter"),
		message: undefinedParamCompCallAsUnknownMsg,
		block:   blockFromRuleName,
	}
}

func notCompParamCompCall() wrapRule {
	return wrapRule{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-param-comp-call", "inline-param-comp-call"),
			not(isInsideRootContent()),
			isNotCompParamCompCall(),
		},
		title:   staticTitle("Not component parameter"),
		message: notCompParamCompCallMsg,
		block:   blockFromRuleName,
	}
}

func alwaysBlock(_ *wrapContext, _ ast.Node) bool { return true }
func neverBlock(_ *wrapContext, _ ast.Node) bool  { return false }
func blockFromRuleName(_ *wrapContext, node ast.Node) bool {
	return strings.HasPrefix(node.Rule().Name(), "block-")
}

func staticTitle(s string) func(*wrapContext, ast.Node) string {
	return func(_ *wrapContext, _ ast.Node) string { return s }
}

func infiniteCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The call to component **" + name + "** creates an infinite loop and was skipped."
}

func infiniteChainCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The call to component **" + name + "** creates an infinite loop and was skipped."
}

func infiniteCompCallByParamMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The call to component **" + name + "** creates an infinite loop and was skipped."
}

func unknownCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The component **" + name + "** is not defined or not registered."
}

func unknownCompParamCallMsg(ctx *wrapContext, compCall ast.Node) string {
	var unknowns []string

	compCallName := getCompCallNameStr(compCall)
	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return "The component **" + compCallName + "** is not defined or not registered."
	}

	resolvedCompArgs := resolveCompArgValues(ctx, compCall)
	paramInfos := getCompDefParamInfos(compDef)

	for _, info := range paramInfos {
		if info.typ != "comp" {
			continue
		}

		value := resolvedCompArgs[info.name]
		if value == "" || strings.HasPrefix(value, "$") || isBuiltInComp(value) {
			continue
		}

		if findCompDef(ctx.root, compCall, value) == nil {
			if !util.InSliceString(value, unknowns) {
				unknowns = append(unknowns, value)
			}
		}
	}

	if len(unknowns) == 0 {
		return "The component **" + compCallName + "** is not defined or not registered."
	}

	if len(unknowns) == 1 {
		return "The component **" + unknowns[0] + "** is not defined or not registered."
	}

	var prepared string
	for i, unk := range unknowns {
		prepared += "**" + unk + "**"
		if i != len(unknowns)-1 {
			prepared += ", "
		}
	}

	return "The components " + prepared + "are not defined or not registered."
}

func blockCompInsideInlineMsg(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "The component **" + name + "** is a block component and cannot be used inline."
}

func blockParamCompInsideInlineMsg(ctx *wrapContext, node ast.Node) string {
	name := getResolvedInlineBlockCompName(ctx, node)
	if name == "" {
		name = getCompCallNameStr(node)
	}
	return "The component **" + name + "** is a block component and cannot be used inline."
}

func undefinedArgMsgFn(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	argNames := getCallArgNames(node)
	return "The parameter(s) **" + strings.Join(argNames, "**, **") + "** are not defined in component **" + name + "**. Only declared parameters can be passed."
}

func wrongArgTypeMsgFn(_ *wrapContext, node ast.Node) string {
	name := getCompCallNameStr(node)
	return "One or more arguments passed to **" + name + "** have the wrong type. Make sure each argument matches the expected type from the component definition."
}

func paramRefInRootMsg(_ *wrapContext, _ ast.Node) string {
	return "Parameters cannot be used in the root context."
}

func paramCompCallInRootMsg(_ *wrapContext, node ast.Node) string {
	name := getParamCompCallNameStr(node)
	return "The parameter component call **{{ $" + name + " }}** cannot be used outside of a component definition."
}

func undefinedParamRefMsg(_ *wrapContext, node ast.Node) string {
	refName := getParamRefNameStr(node)
	return "The parameter **" + refName + "** is not defined for this component."
}

func undefinedParamCompCallAsUnknownMsg(_ *wrapContext, node ast.Node) string {
	name := getParamCompCallNameStr(node)
	return "The parameter **" + name + "** is not defined for this component."
}

func notCompParamCompCallMsg(_ *wrapContext, node ast.Node) string {
	name := getParamCompCallNameStr(node)
	return "The parameter **" + name + "** is not component parameter"
}

func isRuleName(name string) func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return ast.IsRuleName(node, name)
	}
}

func isRuleNameOneOf(names ...string) func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		return ast.IsRuleNameOneOf(node, names)
	}
}

func not(cond func(*wrapContext, ast.Node) bool) func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !cond(ctx, node)
	}
}

func isCalledByItself() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		compCallName := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
		if compCallName == nil {
			return false
		}
		compCallNameStr := strings.TrimSpace(string(compCallName.Raw()))

		compDefs := ast.FilterNodes(ast.GetAncestors(node), func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
		})

		if len(compDefs) == 0 {
			return false
		}

		called := ast.FindNode(compDefs, func(def ast.Node) bool {
			return getCompDefName(def) == compCallNameStr
		})

		return called != nil
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
					return true
				}
			}
		}

		return false
	}
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
				if !ast.IsRuleNameOneOf(n, []string{"block-param-comp-call", "inline-param-comp-call"}) {
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

func isUnknownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compCallName := getCompCallNameStr(node)
		if isBuiltInComp(compCallName) {
			return false
		}
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

		hasCompParam := false
		resolved := resolveCompArgValues(ctx, compCall)

		for _, info := range getCompDefParamInfos(compDef) {
			if info.typ != "comp" {
				continue
			}
			hasCompParam = true

			value := resolved[info.name]
			if value == "" || strings.HasPrefix(value, "$") || isBuiltInComp(value) {
				continue
			}

			if findCompDef(ctx.root, compCall, value) == nil {
				return true
			}
		}

		if !hasCompParam {
			return false
		}

		return false
	}
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

	inlineParamCalls := ast.FilterNodesInTree(compDefContent, func(n ast.Node) bool {
		return ast.IsRuleName(n, "inline-param-comp-call")
	})

	for _, ipc := range inlineParamCalls {
		ipcName := getParamCompCallNameStr(ipc)
		if ipcName == "" {
			continue
		}

		resolvedCompName, ok := resolved[ipcName]
		if !ok || resolvedCompName == "" {
			continue
		}

		if strings.HasPrefix(resolvedCompName, "$") {
			continue
		}

		argCompDef := findCompDef(ctx.root, compCall, resolvedCompName)
		if argCompDef == nil {
			continue
		}

		if isBlockComponent(argCompDef) {
			return resolvedCompName
		}
	}

	return ""
}

func isNotCompParamCompCall() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
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

func hasUndefinedArgs() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		return len(getUndefinedArgNamesWithCtx(ctx, compCall)) > 0
	}
}

func hasWrongTypeArgs() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		if compCallName == "" {
			return false
		}

		compDef := findCompDef(ctx.root, compCall, compCallName)
		if compDef == nil {
			return false
		}

		paramTypeMap := getCompDefParamTypeMap(compDef)
		if isBuiltInComp(compCallName) {
			paramTypeMap = getBuiltInCompParamTypes(compCallName)
		}

		compCallArgsNode := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
		if compCallArgsNode == nil {
			return false
		}

		args := ast.FilterNodes(compCallArgsNode.Children(), func(node ast.Node) bool {
			return ast.IsRuleName(node, "comp-call-arg")
		})

		for _, arg := range args {
			argNameNode := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-name")
			if argNameNode == nil {
				continue
			}
			argNameStr := strings.TrimSpace(string(argNameNode.Raw()))

			expectedType, ok := paramTypeMap[argNameStr]
			if !ok {
				continue
			}

			argType := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-type")
			if argType == nil {
				continue
			}

			actualType := getArgActualType(argType)
			if actualType == "" {
				continue
			}

			if actualType == "param" {
				continue
			}

			if actualType != expectedType {
				return true
			}
		}

		return false
	}
}

func isInsideRootContent() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		ancestors := ast.GetAncestors(node)
		compDef := ast.FindNode(ancestors, func(anc ast.Node) bool {
			return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
		})
		return compDef == nil
	}
}

func isUndefinedParamRef() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, paramRef ast.Node) bool {
		refName := getParamRefNameStr(paramRef)
		if refName == "" {
			return false
		}

		ancestors := ast.GetAncestors(paramRef)
		globalCompDef := ast.FindNode(ancestors, func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "global-comp-def")
		})
		localCompDef := ast.FindNode(ancestors, func(anc ast.Node) bool {
			return ast.IsRuleName(anc, "local-comp-def")
		})

		if globalCompDef == nil && localCompDef == nil {
			return false
		}

		if globalCompDef != nil && localCompDef == nil {
			return !util.InSliceString(refName, getCompDefParamNames(globalCompDef))
		}

		if globalCompDef == nil && localCompDef != nil {
			return !util.InSliceString(refName, getCompDefParamNames(localCompDef))
		}

		if util.InSliceString(refName, getCompDefParamNames(localCompDef)) {
			return false
		}

		return !util.InSliceString(refName, getCompDefParamNames(globalCompDef))
	}
}

func isUndefinedParamCompCall() func(*wrapContext, ast.Node) bool {
	return func(_ *wrapContext, node ast.Node) bool {
		paramName := getParamCompCallNameStr(node)
		if paramName == "" {
			return false
		}

		compDef := findEnclosingCompDef(node)
		if compDef == nil {
			return true
		}

		definedParams := getCompDefParamNames(compDef)

		return !util.InSliceString(paramName, definedParams)
	}
}

func resolveCompArgValues(ctx *wrapContext, compCall ast.Node) map[string]string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return nil
	}

	defCompParams := getCompDefCompParamDefaults(compDef)

	explicitCompArgs := getExplicitCompArgMap(compCall)

	explicitParamArgs := getExplicitParamArgMap(compCall)

	result := make(map[string]string)
	for paramName, defaultValue := range defCompParams {
		if explicit, ok := explicitCompArgs[paramName]; ok {
			result[paramName] = explicit
		} else if paramRef, ok := explicitParamArgs[paramName]; ok {
			result[paramName] = "$" + paramRef
		} else {
			result[paramName] = defaultValue
		}
	}

	return result
}

func getExplicitCompArgMap(compCall ast.Node) map[string]string {
	result := make(map[string]string)

	compCallArgsNode := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgsNode == nil {
		return result
	}

	args := ast.FilterNodes(compCallArgsNode.Children(), func(node ast.Node) bool {
		return ast.IsRuleName(node, "comp-call-arg")
	})

	for _, arg := range args {
		argNameNode := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-name")
		if argNameNode == nil {
			continue
		}
		argNameStr := strings.TrimSpace(string(argNameNode.Raw()))

		argType := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-type")
		if argType == nil {
			continue
		}

		compArg := ast.FindNodeByRuleName(argType.Children(), "comp-call-comp-arg")
		if compArg == nil {
			continue
		}

		argValue := ast.FindNodeByRuleName(compArg.Children(), "comp-call-arg-value")
		if argValue == nil {
			continue
		}

		result[argNameStr] = strings.TrimSpace(string(argValue.Raw()))
	}

	return result
}

func getExplicitParamArgMap(compCall ast.Node) map[string]string {
	result := make(map[string]string)

	compCallArgsNode := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgsNode == nil {
		return result
	}

	args := ast.FilterNodes(compCallArgsNode.Children(), func(node ast.Node) bool {
		return ast.IsRuleName(node, "comp-call-arg")
	})

	for _, arg := range args {
		argNameNode := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-name")
		if argNameNode == nil {
			continue
		}
		argNameStr := strings.TrimSpace(string(argNameNode.Raw()))

		argType := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-type")
		if argType == nil {
			continue
		}

		paramArg := ast.FindNodeByRuleName(argType.Children(), "comp-call-param-arg")
		if paramArg == nil {
			continue
		}

		argValue := ast.FindNodeByRuleName(paramArg.Children(), "comp-call-arg-value")
		if argValue == nil {
			continue
		}

		result[argNameStr] = strings.TrimSpace(string(argValue.Raw()))
	}

	return result
}

func getCompDefCompParamDefaults(compDef ast.Node) map[string]string {
	head := getCompDefHead(compDef)
	result := map[string]string{}

	if head != nil {
		compParamsNode := ast.FindNodeByRuleName(head.Children(), "comp-params")
		if compParamsNode != nil {
			params := ast.FilterNodes(compParamsNode.Children(), func(node ast.Node) bool {
				return ast.IsRuleName(node, "comp-param")
			})

			for _, param := range params {
				nameNode := ast.FindNodeByRuleName(param.Children(), "comp-param-name")
				if nameNode == nil {
					continue
				}
				paramName := strings.TrimSpace(string(nameNode.Raw()))

				typeNode := ast.FindNodeByRuleName(param.Children(), "comp-param-type")
				if typeNode == nil {
					continue
				}

				compCompParam := ast.FindNodeByRuleName(typeNode.Children(), "comp-comp-param")
				if compCompParam == nil {
					continue
				}

				defaValue := ast.FindNodeByRuleName(compCompParam.Children(), "comp-param-defa-value")
				if defaValue == nil {
					continue
				}

				result[paramName] = strings.TrimSpace(string(defaValue.Raw()))
			}
		}
	}

	if len(result) > 0 || !ast.IsRuleName(compDef, "global-comp-def") {
		if len(result) == 0 {
			return nil
		}
		return result
	}

	for _, info := range getCompDefParamInfos(compDef) {
		if info.typ == "comp" && info.defVal != "" {
			result[info.name] = info.defVal
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func getCompDefName(def ast.Node) string {
	if ast.IsRuleName(def, "local-comp-def") {
		head := ast.FindNodeByRuleName(def.Children(), "local-comp-def-head")
		if head == nil {
			return ""
		}
		compName := ast.FindNodeByRuleName(head.Children(), "local-comp-name")
		if compName == nil {
			return ""
		}
		return strings.TrimSpace(string(compName.Raw()))
	}
	if ast.IsRuleName(def, "global-comp-def") {
		compName := ast.FindNodeByRuleName(def.Children(), "global-comp-name")
		if compName == nil {
			return ""
		}
		return strings.TrimSpace(string(compName.Raw()))
	}
	return ""
}

func getCompCallNameStr(node ast.Node) string {
	compCallNameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
	if compCallNameNode != nil {
		return strings.TrimSpace(string(compCallNameNode.Raw()))
	}
	return ""
}

func getParamRefNameStr(node ast.Node) string {
	refNameNode := ast.FindNodeByRuleName(node.Children(), "param-ref-name")
	if refNameNode != nil {
		return strings.TrimSpace(string(refNameNode.Raw()))
	}
	return ""
}

func getParamCompCallNameStr(node ast.Node) string {
	nameNode := ast.FindNodeByRuleName(node.Children(), "param-comp-call-name")
	if nameNode != nil {
		return strings.TrimSpace(string(nameNode.Raw()))
	}
	return ""
}

func getCompDefContent(compDef ast.Node) ast.Node {
	for _, child := range compDef.Children() {
		if child.Rule() == nil {
			continue
		}
		ruleName := child.Rule().Name()
		if ruleName == "local-comp-def-content" || ruleName == "global-comp-def-content" {
			return child
		}
	}
	return nil
}

// TODO: Use ast.GetCompHeadFromCompDef
func getCompDefHead(compDef ast.Node) ast.Node {
	for _, child := range compDef.Children() {
		if child.Rule() == nil {
			continue
		}
		ruleName := child.Rule().Name()
		if ruleName == "local-comp-def-head" || ruleName == "global-comp-def-head" {
			return child
		}
	}
	return nil
}

func getCompDefParamInfos(compDef ast.Node) []compParamInfo {
	compParams := ast.GetCompParamsFromCompDef(compDef)
	if len(compParams) > 0 {
		result := make([]compParamInfo, 0, len(compParams))
		for _, compParam := range compParams {
			if !ast.IsRuleName(compParam, "comp-param") {
				continue
			}

			nameNode := ast.FindNodeByRuleName(compParam.Children(), "comp-param-name")
			if nameNode == nil {
				continue
			}

			typeNode := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
			defVal := ""
			typ := ""
			if typeNode != nil && len(typeNode.Children()) > 0 {
				typ = ast.GetTypeFromCompParam(compParam)
				typeVariant := typeNode.Children()[0]
				defNode := ast.FindNodeByRuleName(typeVariant.Children(), "comp-param-defa-value")
				if defNode != nil {
					defVal = strings.TrimSpace(string(defNode.Raw()))
				}
			}

			result = append(result, compParamInfo{
				name:   strings.TrimSpace(string(nameNode.Raw())),
				typ:    typ,
				defVal: defVal,
			})
		}
		return result
	}

	// Fallback for global components whose header params were not parsed into AST.
	if !ast.IsRuleName(compDef, "global-comp-def") {
		return nil
	}

	raw := strings.TrimSpace(string(compDef.Raw()))
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	result := []compParamInfo{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := globalCompParamLineRE.FindStringSubmatch(line)
		if matches == nil {
			break
		}

		name := matches[1]
		value := matches[2]
		typ := "string"

		switch {
		case strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\""):
			typ = "string"
		case value == "true" || value == "false":
			typ = "bool"
		case numberValueRE.MatchString(value):
			typ = "number"
		case compValueRE.MatchString(value):
			typ = "comp"
		}

		result = append(result, compParamInfo{
			name:   name,
			typ:    typ,
			defVal: value,
		})
	}

	return result
}

func getCompDefParamNames(compDef ast.Node) []string {
	head := getCompDefHead(compDef)
	var names []string

	if head != nil {
		compParamsNode := ast.FindNodeByRuleName(head.Children(), "comp-params")
		if compParamsNode != nil {
			params := ast.FilterNodes(compParamsNode.Children(), func(node ast.Node) bool {
				return ast.IsRuleName(node, "comp-param")
			})

			for _, param := range params {
				nameNode := ast.FindNodeByRuleName(param.Children(), "comp-param-name")
				if nameNode == nil {
					continue
				}
				names = append(names, strings.TrimSpace(string(nameNode.Raw())))
			}
		}
	}

	if len(names) > 0 || !ast.IsRuleName(compDef, "global-comp-def") {
		return names
	}

	for _, info := range getCompDefParamInfos(compDef) {
		names = append(names, info.name)
	}
	return names
}

func getCompDefParamTypeMap(compDef ast.Node) map[string]string {
	head := getCompDefHead(compDef)
	result := map[string]string{}

	if head != nil {
		compParamsNode := ast.FindNodeByRuleName(head.Children(), "comp-params")
		if compParamsNode != nil {
			params := ast.FilterNodes(compParamsNode.Children(), func(node ast.Node) bool {
				return ast.IsRuleName(node, "comp-param")
			})

			for _, param := range params {
				nameNode := ast.FindNodeByRuleName(param.Children(), "comp-param-name")
				if nameNode == nil {
					continue
				}
				paramName := strings.TrimSpace(string(nameNode.Raw()))

				typeNode := ast.FindNodeByRuleName(param.Children(), "comp-param-type")
				if typeNode == nil {
					continue
				}

				result[paramName] = getParamDefType(typeNode)
			}
		}
	}

	if len(result) > 0 || !ast.IsRuleName(compDef, "global-comp-def") {
		if len(result) == 0 {
			return nil
		}
		return result
	}

	for _, info := range getCompDefParamInfos(compDef) {
		result[info.name] = info.typ
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func getParamDefType(typeNode ast.Node) string {
	if typeNode == nil {
		return ""
	}
	if ast.FindNodeByRuleName(typeNode.Children(), "comp-string-param") != nil {
		return "string"
	}
	if ast.FindNodeByRuleName(typeNode.Children(), "comp-number-param") != nil {
		return "number"
	}
	if ast.FindNodeByRuleName(typeNode.Children(), "comp-bool-param") != nil {
		return "bool"
	}
	if ast.FindNodeByRuleName(typeNode.Children(), "comp-comp-param") != nil {
		return "comp"
	}
	return ""
}

func getArgActualType(argTypeNode ast.Node) string {
	if ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-string-arg") != nil {
		return "string"
	}
	if ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-number-arg") != nil {
		return "number"
	}
	if ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-bool-arg") != nil {
		return "bool"
	}
	if ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-comp-arg") != nil {
		return "comp"
	}
	if ast.FindNodeByRuleName(argTypeNode.Children(), "comp-call-param-arg") != nil {
		return "param"
	}
	return ""
}

func getCallArgNames(compCall ast.Node) []string {
	compCallArgsNode := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgsNode == nil {
		return nil
	}

	args := ast.FilterNodes(compCallArgsNode.Children(), func(node ast.Node) bool {
		return ast.IsRuleName(node, "comp-call-arg")
	})

	var names []string
	for _, arg := range args {
		argNameNode := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-name")
		if argNameNode == nil {
			continue
		}
		names = append(names, strings.TrimSpace(string(argNameNode.Raw())))
	}
	return names
}

func getUndefinedArgNamesWithCtx(ctx *wrapContext, compCall ast.Node) []string {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return nil
	}

	compDef := findCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return nil
	}

	definedParams := getCompDefParamNames(compDef)
	if isBuiltInComp(compCallName) {
		definedParams = getBuiltInCompParams(compCallName)
	}

	compCallArgsNode := ast.FindNodeByRuleName(compCall.Children(), "comp-call-args")
	if compCallArgsNode == nil {
		return nil
	}

	args := ast.FilterNodes(compCallArgsNode.Children(), func(node ast.Node) bool {
		return ast.IsRuleName(node, "comp-call-arg")
	})

	var undefs []string
	for _, arg := range args {
		argNameNode := ast.FindNodeByRuleName(arg.Children(), "comp-call-arg-name")
		if argNameNode == nil {
			continue
		}
		argNameStr := strings.TrimSpace(string(argNameNode.Raw()))
		if !util.InSliceString(argNameStr, definedParams) {
			undefs = append(undefs, argNameStr)
		}
	}
	return undefs
}

func findEnclosingCompDef(node ast.Node) ast.Node {
	ancestors := ast.GetAncestors(node)
	return ast.FindNode(ancestors, func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
	})
}

func isBlockComponent(compDef ast.Node) bool {
	compDefContent := getCompDefContent(compDef)
	if compDefContent == nil {
		return false
	}

	childrenCount := len(compDefContent.Children())
	if childrenCount == 0 {
		return false
	} else if childrenCount > 1 {
		return true
	}

	p := ast.FindNodeByRuleName(compDefContent.Children(), "p")
	if p == nil {
		return true
	}

	pContent := ast.FindNodeByRuleName(p.Children(), "p-content")
	softBlock := ast.FindNodeByRuleName(pContent.Children(), "soft-break")

	if softBlock != nil {
		return true
	}

	return false
}

func findCompDef(root ast.Node, compCallNode ast.Node, name string) ast.Node {
	globalCompDefAnc := ast.FindNode(ast.GetAncestors(compCallNode), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "global-comp-def")
	})

	localCompDefSrc := root
	if globalCompDefAnc != nil {
		localCompDefSrc = globalCompDefAnc
	}

	localCompDef := ast.FindLocalCompDef(localCompDefSrc, name)
	if localCompDef != nil {
		return localCompDef
	}

	globalCompDef := ast.FindGlobalCompDef(root, name)
	if globalCompDef != nil {
		return globalCompDef
	}

	return nil
}

func isBuiltInComp(name string) bool {
	return util.InSliceString(name, []string{"LINK"})
}

func getBuiltInCompParams(name string) []string {
	switch name {
	case "LINK":
		return []string{"url", "text", "new-tab"}
	}
	return nil
}

func getBuiltInCompParamTypes(name string) map[string]string {
	switch name {
	case "LINK":
		return map[string]string{
			"url":     "string",
			"text":    "string",
			"new-tab": "bool",
		}
	}
	return nil
}
