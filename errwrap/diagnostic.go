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

func (a conditionAnalyzer) Diagnose(ctx *wrapContext, node ast.Node) (diagnostic, bool) {
	for _, cond := range a.conditions {
		if !cond(ctx, node) {
			return diagnostic{}, false
		}
	}

	return diagnostic{
		title:   a.title(ctx, node),
		message: a.message(ctx, node),
		block:   a.block(ctx, node),
	}, true
}
