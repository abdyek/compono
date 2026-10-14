package compono

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

// analysisRefs returns the references to components under top in source
// order, and the index of every reference node in the returned list.
func analysisRefs(source []byte, top ast.Node) ([]Ref, map[ast.Node]int) {
	var refs []Ref
	index := map[ast.Node]int{}

	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if ref, ok := analysisRef(source, top, node); ok {
			index[node] = len(refs)
			refs = append(refs, ref)
		}
		for _, child := range node.Children() {
			walk(child)
		}
	}
	walk(top)

	return refs, index
}

// analysisRef returns the reference node is, if it is one.
func analysisRef(source []byte, top, node ast.Node) (Ref, bool) {
	var ref Ref
	switch {
	case ast.IsRuleNameOneOf(node, []string{"block-comp-call", "inline-comp-call"}):
		nameNode := ast.FindNodeByRuleName(node.Children(), "comp-call-name")
		if nameNode == nil {
			return Ref{}, false
		}
		ref = Ref{
			Name:  strings.TrimSpace(string(nameNode.Raw())),
			Kind:  "call",
			Args:  analysisArgNames(ast.FindNodeByRuleName(node.Children(), "comp-call-args")),
			Block: ast.IsRuleName(node, "block-comp-call"),
		}
	case ast.IsRuleNameOneOf(node, []string{"comp-call-comp-arg", "comp-comp-param"}):
		nameNode := ast.FindNode(node.Children(), func(child ast.Node) bool {
			return ast.IsRuleNameOneOf(child, []string{"comp-call-arg-value", "comp-param-defa-value"})
		})
		if nameNode == nil {
			return Ref{}, false
		}
		ref = Ref{
			Name: strings.TrimSpace(string(nameNode.Raw())),
			Kind: "value",
			Args: analysisArgNames(ast.FindNodeByRuleName(node.Children(), "comp-bound-args")),
		}
		if ast.FindNodeByRuleName(ast.GetAncestors(node), "comp-param") != nil {
			ref.Kind = "default"
		}
	default:
		return Ref{}, false
	}

	ref.In = enclosingLocalName(node)
	ref.Local = ast.FindLocalCompDef(top, ref.Name) != nil
	ref.Range = trimmedRange(source, node.Range())
	return ref, true
}

// analysisArgNames returns the names of the arguments under args, nil if
// there are none.
func analysisArgNames(args ast.Node) []string {
	if args == nil {
		return nil
	}
	var names []string
	for _, arg := range args.Children() {
		if ast.IsRuleName(arg, "comp-call-arg") {
			names = append(names, ast.GetArgNameFromCompCallArg(arg))
		}
	}
	return names
}
