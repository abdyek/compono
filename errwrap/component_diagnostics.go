package errwrap

import (
	"bytes"
	"strings"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/selector"
	"github.com/umono-cms/compono/util"
)

func invalidParamDef() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("block-comp-call", "inline-comp-call"),
			isKnownComponent(),
			hasInvalidParamDef(),
		},
		title:   staticTitle("Invalid parameter definition"),
		message: invalidParamDefMsg,
		block:   blockFromRuleName,
	}
}

func hasInvalidParamDef() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		_, ok := getInvalidCompParamDef(ctx, compCall)
		return ok
	}
}

// getInvalidCompParamDef returns the first invalid parameter definition of the
// local or global component definition the call resolves to. For a local
// definition the scanned text is the definition head after the component name;
// for a global definition it is the definition head after its first "~".
// An indented line right after the definition line is also invalid.
func getInvalidCompParamDef(ctx *wrapContext, compCall ast.Node) (string, bool) {
	compCallName := getCompCallNameStr(compCall)
	if compCallName == "" {
		return "", false
	}

	compDef := ast.FindCompDef(ctx.root, compCall, compCallName)
	if compDef == nil {
		return "", false
	}

	return getInvalidParamDefOfCompDef(compDef)
}

// getInvalidParamDefOfCompDef returns the first invalid parameter definition
// of a local or global component definition. It returns false for any other
// node, such as a built-in.
func getInvalidParamDefOfCompDef(compDef ast.Node) (string, bool) {
	if ast.IsRuleName(compDef, "global-comp-def") {
		return getInvalidGlobalParamLine(compDef)
	}

	if !ast.IsRuleName(compDef, "local-comp-def") {
		return "", false
	}

	head := ast.FindNodeByRuleName(compDef.Children(), "local-comp-def-head")
	if head == nil {
		return "", false
	}

	nameNode := ast.FindNodeByRuleName(head.Children(), "local-comp-name")
	if nameNode == nil {
		return "", false
	}
	name := strings.TrimSpace(string(nameNode.Raw()))
	if name == "" {
		return "", false
	}

	headRaw := string(head.Raw())
	nameIdx := strings.Index(headRaw, name)
	if nameIdx < 0 {
		return "", false
	}

	if invalid, ok := selector.FirstInvalidCompParamDef([]byte(headRaw[nameIdx+len(name):])); ok {
		return invalid, true
	}
	return indentedLineAfterDefLine(compDef.Raw())
}

// getInvalidGlobalParamLine returns the first invalid parameter definition found
// on the global component definition head after its leading "~".
func getInvalidGlobalParamLine(compDef ast.Node) (string, bool) {
	head := ast.FindNodeByRuleName(compDef.Children(), "global-comp-def-head")
	if head == nil {
		return "", false
	}

	headRaw := string(head.Raw())
	tildeIdx := strings.Index(headRaw, "~")
	if tildeIdx < 0 {
		return "", false
	}

	tail := headRaw[tildeIdx+1:]
	if strings.Trim(tail, " \t\n") == "" {
		return "~", true
	}

	if invalid, ok := selector.FirstInvalidCompParamDef([]byte(tail)); ok {
		return invalid, true
	}
	return indentedLineAfterDefLine(compDef.Raw())
}

// indentedLineAfterDefLine returns the trimmed text of the first indented line
// immediately following the definition's "~" line, when one exists.
func indentedLineAfterDefLine(compDefRaw []byte) (string, bool) {
	tildeIdx := bytes.IndexByte(compDefRaw, '~')
	if tildeIdx < 0 {
		return "", false
	}
	_, rest, ok := bytes.Cut(compDefRaw[tildeIdx:], []byte("\n"))
	if !ok {
		return "", false
	}
	line, _, _ := bytes.Cut(rest, []byte("\n"))
	trimmed := bytes.Trim(line, " \t")
	if len(trimmed) == 0 || (line[0] != ' ' && line[0] != '\t') {
		return "", false
	}
	return string(trimmed), true
}

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
		return ast.FindCompDef(ctx.root, node, compCallName) == nil
	}
}

func isKnownComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		return !isUnknownComponent()(ctx, node)
	}
}

func isNotBuiltinComponent() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, node ast.Node) bool {
		compDef := ast.FindCompDef(ctx.root, node, getCompCallNameStr(node))
		return !ast.IsRuleName(compDef, "builtin-comp")
	}
}

func hasUnknownResolvedCompArg() func(*wrapContext, ast.Node) bool {
	return func(ctx *wrapContext, compCall ast.Node) bool {
		compCallName := getCompCallNameStr(compCall)
		compDef := ast.FindCompDef(ctx.root, compCall, compCallName)
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
	for _, paramCompCall := range ast.FilterNodesInDefContent(compDefContent, func(node ast.Node) bool {
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
		if _, ok := explicitCompArgs[info.name]; !ok {
			continue
		}

		value := resolvedCompArgs[info.name]
		if value == "" || strings.HasPrefix(value, "$") {
			continue
		}

		if ast.FindCompDef(ctx.root, compCall, value) == nil && !util.InSliceString(value, unknowns) {
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

		compDef := ast.FindCompDef(ctx.root, node, compCallName)
		if compDef == nil {
			return false
		}

		return isBlockComponent(compDef)
	}
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
