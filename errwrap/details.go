package errwrap

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

// detailsValueError returns the error of a call to the DETAILS built-in whose
// parameters resolved to values.
func detailsValueError(root, call ast.Node, invokerAncestors []ast.Node) (title, message string, ok bool) {
	if util.TrimWhitespace(ast.ResolveFrameParam(root, call, "text", invokerAncestors).Raw) == "" {
		return "Empty text", "The parameter **text** cannot be empty.", true
	}
	return "", "", false
}
