package errwrap

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

type compParamInfo struct {
	name   string
	typ    string
	defVal string
}

func getCompCallNameStr(node ast.Node) string {
	compCallNameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
	if compCallNameNode != nil {
		return strings.TrimSpace(string(compCallNameNode.Raw()))
	}
	return ""
}

func getParamRefNameStr(node ast.Node) string {
	return ast.GetParamRefName(node)
}

func getParamCompCallNameStr(node ast.Node) string {
	nameNode := ast.FindNodeByRuleName(node.Children(), "param-comp-call-name")
	if nameNode != nil {
		return strings.TrimSpace(string(nameNode.Raw()))
	}
	return getParamRefNameStr(node)
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

func getCompDefParamInfos(compDef ast.Node) []compParamInfo {
	compParams := ast.GetCompParamsFromCompDef(compDef)
	if len(compParams) == 0 {
		return nil
	}

	result := make([]compParamInfo, 0, len(compParams))
	for _, compParam := range compParams {
		if !ast.IsRuleName(compParam, "comp-param") {
			continue
		}

		name := ast.GetParamNameFromCompParam(compParam)
		if name == "" {
			continue
		}

		defVal := ""
		compParamType := ast.FindNodeByRuleName(compParam.Children(), "comp-param-type")
		typ := "comp"
		if compParamType != nil && len(compParamType.Children()) == 0 {
			continue
		}
		if compParamType != nil && len(compParamType.Children()) > 0 {
			typeVariant := compParamType.Children()[0]
			typ = ast.GetTypeFromCompParam(compParam)
			if typ == "context" {
				root := compDef
				ancestors := ast.GetAncestors(compDef)
				if len(ancestors) > 0 {
					root = ancestors[len(ancestors)-1]
				}
				typ = ast.ResolveCompParamDefaultFromCompDef(root, compDef, name).Type
			}
			if ast.FindNodeByRuleName(typeVariant.Children(), "comp-param-defa-value") != nil {
				defVal = ast.GetParamDefValFromCompParam(compParam)
			}
		}

		result = append(result, compParamInfo{
			name:   name,
			typ:    typ,
			defVal: defVal,
		})
	}

	return result
}

func getCompDefParamNames(compDef ast.Node) []string {
	names := make([]string, 0)
	for _, info := range getCompDefParamInfos(compDef) {
		names = append(names, info.name)
	}
	return names
}

func getCompDefParamTypeMap(compDef ast.Node) map[string]string {
	result := map[string]string{}
	for _, info := range getCompDefParamInfos(compDef) {
		result[info.name] = info.typ
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func findEnclosingCompDef(node ast.Node) ast.Node {
	ancestors := ast.GetAncestors(node)
	return ast.FindNode(ancestors, func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{"local-comp-def", "global-comp-def"})
	})
}

func isInlineParamRefNode(node ast.Node) bool {
	if !ast.IsRuleName(node, "param-ref") {
		return false
	}

	pContent := ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleName(anc, "p-content")
	})
	if pContent != nil {
		hasSoftBreak := false
		for _, child := range pContent.Children() {
			if ast.IsRuleName(child, "soft-break") {
				hasSoftBreak = true
			}
		}
		if hasSoftBreak {
			return false
		}

		for _, child := range pContent.Children() {
			if child == node {
				continue
			}
			if ast.IsRuleName(child, "plain") && strings.TrimSpace(string(child.Raw())) == "" {
				continue
			}
			return true
		}
		return false
	}

	return ast.FindNode(ast.GetAncestors(node), func(anc ast.Node) bool {
		return ast.IsRuleNameOneOf(anc, []string{
			"h1-content",
			"h2-content",
			"h3-content",
			"h4-content",
			"h5-content",
			"h6-content",
			"em-content",
			"strong-content",
			"link-text",
		})
	}) != nil
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

	return softBlock != nil
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

	builtinCompDef := ast.FindBuiltinCompDef(root, name)
	if builtinCompDef != nil {
		return builtinCompDef
	}

	return nil
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
	return getExplicitArgMapByType(compCall, "comp")
}

func getExplicitParamArgMap(compCall ast.Node) map[string]string {
	return getExplicitArgMapByType(compCall, "param")
}

func getExplicitArgMapByType(compCall ast.Node, typ string) map[string]string {
	result := make(map[string]string)

	for _, arg := range ast.GetCompCallArgsFromCompCall(compCall) {
		if !ast.IsRuleName(arg, "comp-call-arg") {
			continue
		}
		if ast.GetTypeFromCompCallArg(arg) != typ {
			continue
		}
		result[ast.GetArgNameFromCompCallArg(arg)] = ast.GetArgValueFromCompCallArg(arg)
	}

	return result
}

func getCompDefCompParamDefaults(compDef ast.Node) map[string]string {
	result := map[string]string{}
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
