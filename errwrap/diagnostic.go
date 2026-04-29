package errwrap

import "github.com/umono-cms/compono/ast"

type diagnostic struct {
	title   string
	message string
	block   bool
}

type diagnosticAnalyzer interface {
	Diagnose(ctx *wrapContext, node ast.Node) (diagnostic, bool)
}

func (wr wrapRule) Diagnose(ctx *wrapContext, node ast.Node) (diagnostic, bool) {
	for _, cond := range wr.conditions {
		if !cond(ctx, node) {
			return diagnostic{}, false
		}
	}

	return diagnostic{
		title:   wr.title(ctx, node),
		message: wr.message(ctx, node),
		block:   wr.block(ctx, node),
	}, true
}
