package errwrap

import (
	"github.com/umono-cms/compono/ast"
)

// listValueError returns the error of a call to the LIST built-in whose item
// component cannot be used. It does not depend on items.
func listValueError(root, call ast.Node, invokerAncestors []ast.Node) (title, message string, ok bool) {
	value := ast.ResolveFrameParam(root, call, "item", invokerAncestors)
	if value.Type != "comp" || value.Raw == "" {
		return "", "", false
	}

	compDef := ast.FindCompDefInScope(root, value.Scope, call, value.Raw)
	if compDef == nil {
		return "", "", false
	}

	if ast.IsRuleName(compDef, "builtin-comp") {
		return "Invalid item component", "The component **" + value.Raw + "** must define a record parameter named **item**.", true
	}

	hasItemRecord := false
	for _, info := range getCompDefParamInfos(compDef) {
		if info.name == "item" && info.typ == "record" {
			hasItemRecord = true
			break
		}
	}
	if !hasItemRecord {
		return "Invalid item component", "The component **" + value.Raw + "** must define a record parameter named **item**.", true
	}

	if value.Bound.Arg("item") != nil {
		return "Duplicate argument", duplicateArgumentMsg("item", value.Raw), true
	}

	missing := []string{}
	for _, compParam := range ast.GetCompParamsFromCompDef(compDef) {
		if !ast.IsRuleName(compParam, "comp-param") || !ast.IsRequiredCompParam(compParam) {
			continue
		}
		name := ast.GetParamNameFromCompParam(compParam)
		if name == "" || name == "item" {
			continue
		}
		if value.Bound.Arg(name) == nil {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return "Missing argument", missingArgNamesMsg(value.Raw, missing), true
	}

	return "", "", false
}
