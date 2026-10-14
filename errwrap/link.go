package errwrap

import (
	"github.com/umono-cms/compono/ast"
	"github.com/umono-cms/compono/util"
)

// linkCurrentValues are the values allowed for the LINK "current" parameter.
var linkCurrentValues = []string{"page", "step", "location", "date", "time", "true"}

// linkValueError returns the error of a call to the LINK built-in whose
// parameters resolved to values.
func linkValueError(root, call ast.Node, invokerAncestors []ast.Node) (title, message string, ok bool) {
	text := ast.ResolveFrameParam(root, call, "text", invokerAncestors).Raw
	url := ast.ResolveFrameParam(root, call, "url", invokerAncestors).Raw
	current := ast.ResolveFrameParam(root, call, "current", invokerAncestors).Raw

	if util.TrimWhitespace(text) == "" {
		return "Empty text", "The parameter **text** cannot be empty.", true
	}

	if util.NormalizeURL(url) == "" {
		return "Empty URL", "The parameter **url** cannot be empty.", true
	}

	if scheme, disallowed := util.DisallowedURLScheme(url); disallowed {
		title, message := DisallowedURLSchemeError(scheme)
		return title, message, true
	}

	current = util.TrimWhitespace(current)
	if current != "" && !util.InSliceString(current, linkCurrentValues) {
		return "Invalid current value", "The value **" + current + "** is not allowed for the parameter **current**. Allowed values: **page**, **step**, **location**, **date**, **time**, **true**.", true
	}

	return "", "", false
}

// DisallowedURLSchemeError returns the error of a LINK or IMAGE argument
// or a markdown link URL whose scheme is not http, https, mailto or tel.
func DisallowedURLSchemeError(scheme string) (title, message string) {
	return "Disallowed URL scheme", "The URL scheme **" + scheme + "** is not allowed. Allowed schemes: **http**, **https**, **mailto**, **tel**."
}

// isCompCallInsideLink reports whether node is a component call in the text
// or URL of a markdown link. The link renderer reports such a call as
// invalid-component-usage and drops the link, so the call and its arguments
// are not diagnosed.
func isCompCallInsideLink(node ast.Node) bool {
	if !ast.IsRuleName(node, "inline-comp-call") {
		return false
	}

	for _, anc := range ast.GetAncestors(node) {
		if ast.IsRuleName(anc, "link-text") || ast.IsRuleName(anc, "link-url") {
			return true
		}
	}

	return false
}
