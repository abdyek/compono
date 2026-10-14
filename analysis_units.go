package compono

import (
	"strings"

	"github.com/umono-cms/compono/ast"
)

// markdownElementOfRule maps the rule names of Markdown element nodes to
// the names of MarkdownElements.
var markdownElementOfRule = map[string]string{
	"h1":          "heading",
	"h2":          "heading",
	"h3":          "heading",
	"h4":          "heading",
	"h5":          "heading",
	"h6":          "heading",
	"p":           "p",
	"soft-break":  "br",
	"em":          "em",
	"strong":      "strong",
	"link":        "link",
	"inline-code": "inline-code",
	"code-block":  "code-block",
	"list":        "list",
	"blockquote":  "blockquote",
	"hr":          "hr",
	"table":       "table",
}

// analysisElements returns the Markdown elements under top in source order.
func analysisElements(source []byte, top ast.Node) []Element {
	var elements []Element
	var walk func(node ast.Node)
	walk = func(node ast.Node) {
		if node.Rule() != nil {
			if name, ok := markdownElementOfRule[node.Rule().Name()]; ok {
				elements = append(elements, Element{
					Name:  name,
					In:    enclosingLocalName(node),
					Range: trimmedRange(source, node.Range()),
				})
			}
		}
		for _, child := range node.Children() {
			walk(child)
		}
	}
	walk(top)
	return elements
}

// analysisRoot returns the units of the content of top. A block call is a
// call; a paragraph is split into its inline calls, its inline Markdown
// elements and its text; any other block is Markdown, except an invalid
// unit, which is text. Line breaks and text of only spaces are not units.
// refs maps reference nodes to their index in Refs.
func analysisRoot(source []byte, top ast.Node, refs map[ast.Node]int) []Unit {
	content := ast.FindNode(top.Children(), func(child ast.Node) bool {
		return ast.IsRuleNameOneOf(child, []string{"root-content", "global-comp-def-content"})
	})
	if content == nil {
		return nil
	}

	var units []Unit
	add := func(kind UnitKind, node ast.Node) {
		ref := -1
		if kind == "call" {
			ref = refs[node]
		}
		units = append(units, Unit{Kind: kind, Ref: ref, Range: trimmedRange(source, node.Range())})
	}

	for _, block := range content.Children() {
		switch {
		case ast.IsRuleName(block, "block-comp-call"):
			add("call", block)
		case ast.IsRuleName(block, "block-invalid-expression"):
			add("text", block)
		case ast.IsRuleName(block, "p"):
			pContent := ast.FindNodeByRuleName(block.Children(), "p-content")
			if pContent == nil {
				continue
			}
			for _, inline := range pContent.Children() {
				switch {
				case ast.IsRuleName(inline, "soft-break"):
				case ast.IsRuleName(inline, "plain") && strings.TrimSpace(string(inline.Raw())) == "":
				case ast.IsRuleName(inline, "inline-comp-call"):
					add("call", inline)
				case ast.IsRuleNameOneOf(inline, []string{"em", "strong", "link", "inline-code"}):
					add("markdown", inline)
				default:
					add("text", inline)
				}
			}
		default:
			add("markdown", block)
		}
	}
	return units
}
