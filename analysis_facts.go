package compono

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

// trimmedRange returns rng in source without the spaces, tabs and line
// breaks at its ends. A range of only such bytes is returned as it is.
func trimmedRange(source []byte, rng ast.Range) Range {
	start, end := rng.Start, rng.End
	for start < end && isAnalysisSpace(source[start]) {
		start++
	}
	for end > start && isAnalysisSpace(source[end-1]) {
		end--
	}
	if start == end {
		start, end = rng.Start, rng.End
	}
	return Range{Start: positionIn(source, start), End: positionIn(source, end)}
}

func isAnalysisSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// enclosingLocalName returns the name of the local definition node is in,
// or "" outside local definitions.
func enclosingLocalName(node ast.Node) string {
	for cur := node; cur != nil; cur = cur.Parent() {
		if !ast.IsRuleName(cur, "local-comp-def") {
			continue
		}
		head := ast.FindNodeByRuleName(cur.Children(), "local-comp-def-head")
		if head == nil {
			return ""
		}
		name := ast.FindNodeByRuleName(head.Children(), "local-comp-name")
		if name == nil {
			return ""
		}
		return strings.TrimSpace(string(name.Raw()))
	}
	return ""
}

// analysisParams returns the parameter definitions under compParams, nil if
// there are none. A definition without a type is skipped.
func analysisParams(source []byte, compParams ast.Node) []Param {
	if compParams == nil {
		return nil
	}
	var params []Param
	for _, cp := range compParams.Children() {
		if !ast.IsRuleName(cp, "comp-param") {
			continue
		}
		typeNode := ast.FindNodeByRuleName(cp.Children(), "comp-param-type")
		if typeNode == nil || len(typeNode.Children()) == 0 {
			continue
		}
		typ := strings.TrimSuffix(strings.TrimPrefix(typeNode.Children()[0].Rule().Name(), "comp-"), "-param")
		if typ == "comp" {
			typ = "component"
		}
		params = append(params, Param{
			Name:     strings.TrimSuffix(ast.GetParamNameFromCompParam(cp), "!"),
			Type:     ParamType(typ),
			Required: ast.IsRequiredCompParam(cp),
			Range:    trimmedRange(source, cp.Range()),
		})
	}
	return params
}

// analysisLocals returns the local definitions of top, nil if there are none.
func analysisLocals(source []byte, top ast.Node) []Local {
	wrapper := ast.FindNodeByRuleName(top.Children(), "local-comp-def-wrapper")
	if wrapper == nil {
		return nil
	}
	var locals []Local
	for _, def := range wrapper.Children() {
		if !ast.IsRuleName(def, "local-comp-def") {
			continue
		}
		local := Local{Range: trimmedRange(source, def.Range())}
		if head := ast.FindNodeByRuleName(def.Children(), "local-comp-def-head"); head != nil {
			if name := ast.FindNodeByRuleName(head.Children(), "local-comp-name"); name != nil {
				local.Name = strings.TrimSpace(string(name.Raw()))
			}
			local.Params = analysisParams(source, ast.FindNodeByRuleName(head.Children(), "comp-params"))
		}
		locals = append(locals, local)
	}
	return locals
}
