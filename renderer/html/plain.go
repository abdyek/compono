package html

import (
	"html"
	"regexp"

	"github.com/umono-cms/compono/ast"
)

type plain struct {
	baseRenderable
	renderer *renderer
}

func newPlain(rend *renderer) renderableNode {
	return &plain{
		renderer: rend,
	}
}

func (p *plain) New() renderableNode {
	return newPlain(p.renderer)
}

func (_ *plain) Condition(invoker renderableNode, node ast.Node) bool {
	return ast.IsRuleName(node, "plain")
}

func (p *plain) Render() string {
	return html.EscapeString(p.normalizeEdges(string(p.Node().Raw())))
}

func (_ *plain) normalizeEdges(s string) string {
	reJunk := regexp.MustCompile(`[\t\n\r\f\v]+`)
	s = reJunk.ReplaceAllString(s, "")

	re := regexp.MustCompile(`^\s{2,}|\s{2,}$`)
	return re.ReplaceAllString(s, " ")
}
