package compono

import (
	"bytes"

	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/errwrap"
)

// analysisDefinitionDiagnostics returns an invalid-parameter-definition for
// the parameter line of a global source and for every local definition of
// top whose parameter definitions are invalid, at the invalid text. They
// are reported whether or not the definition is called.
func analysisDefinitionDiagnostics(source []byte, top ast.Node) []Diagnostic {
	defs := []ast.Node{}
	if ast.IsRuleName(top, "global-comp-def") {
		defs = append(defs, top)
	}
	if wrapper := ast.FindNodeByRuleName(top.Children(), "local-comp-def-wrapper"); wrapper != nil {
		for _, def := range wrapper.Children() {
			if ast.IsRuleName(def, "local-comp-def") {
				defs = append(defs, def)
			}
		}
	}

	var diags []Diagnostic
	for _, def := range defs {
		text, ok := errwrap.InvalidParamDefinition(def)
		if !ok {
			continue
		}
		message := "The parameter definition **" + text + "** is invalid."
		if name := enclosingLocalName(def); name != "" {
			message = "The parameter definition **" + text + "** of component **" + name + "** is invalid."
		}
		diags = append(diags, Diagnostic{
			Code:    CodeInvalidParameterDefinition,
			Message: message,
			Range:   trimmedRange(source, invalidDefinitionRange(source, def, text)),
		})
	}
	return diags
}

// invalidDefinitionRange returns the range of text in def: its first
// occurrence from the head of def on that is outside every parameter
// definition that has a type. It returns the range of def when there is none.
func invalidDefinitionRange(source []byte, def ast.Node, text string) ast.Range {
	head := ast.FindNode(def.Children(), func(child ast.Node) bool {
		return ast.IsRuleNameOneOf(child, []string{"local-comp-def-head", "global-comp-def-head"})
	})
	start, end := def.Range().Start, def.Range().End
	var valid []ast.Range
	if head != nil {
		start = head.Range().Start
		if params := ast.FindNodeByRuleName(head.Children(), "comp-params"); params != nil {
			for _, param := range params.Children() {
				typeNode := ast.FindNodeByRuleName(param.Children(), "comp-param-type")
				if typeNode == nil || len(typeNode.Children()) == 0 {
					continue
				}
				valid = append(valid, param.Range())
			}
		}
	}

	for from := start; from < end; {
		idx := bytes.Index(source[from:end], []byte(text))
		if idx < 0 || text == "" {
			break
		}
		s := from + idx
		e := s + len(text)
		overlaps := false
		for _, v := range valid {
			if s < v.End && v.Start < e {
				overlaps = true
				break
			}
		}
		if !overlaps {
			return ast.Range{Start: s, End: e}
		}
		from = s + 1
	}
	return def.Range()
}
