package rule

import (
	"github.com/umono-cms/compono/selector"
)

// invalidExpression is a {{ }} unit in inline content that follows no grammar.
type invalidExpression struct{}

func newInvalidExpression() Rule {
	return &invalidExpression{}
}

func (_ *invalidExpression) Name() string {
	return "invalid-expression"
}

func (_ *invalidExpression) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewInvalidExpressions(false),
	}
}

func (_ *invalidExpression) Rules() []Rule {
	return []Rule{}
}

// blockInvalidExpression is a {{ }} unit that follows no grammar and stands on its own lines.
type blockInvalidExpression struct{}

func newBlockInvalidExpression() Rule {
	return &blockInvalidExpression{}
}

func (_ *blockInvalidExpression) Name() string {
	return "block-invalid-expression"
}

func (_ *blockInvalidExpression) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewInvalidExpressions(true),
	}
}

func (_ *blockInvalidExpression) Rules() []Rule {
	return []Rule{}
}

// unclosedExpression is a {{ without a matching }}.
type unclosedExpression struct{}

func newUnclosedExpression() Rule {
	return &unclosedExpression{}
}

func (_ *unclosedExpression) Name() string {
	return "unclosed-expression"
}

func (_ *unclosedExpression) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewUnclosedExpressions(),
	}
}

func (_ *unclosedExpression) Rules() []Rule {
	return []Rule{}
}
