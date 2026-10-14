package errwrap

import "github.com/umono-cms/compono/ast"

// invalidExpression reports a {{ }} unit that follows no grammar. The unit
// drops as a whole and the names inside it are not diagnosed.
func invalidExpression() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("invalid-expression", "block-invalid-expression"),
		},
		title:   staticTitle("Invalid expression"),
		message: staticTitle("The content between **{{** and **}}** must be a component call, a parameter reference or a context reference."),
		block:   blockFromRuleName,
	}
}

// unclosedExpression reports a {{ without a matching }}. Only the {{ drops.
func unclosedExpression() conditionAnalyzer {
	return conditionAnalyzer{
		conditions: []func(*wrapContext, ast.Node) bool{
			isRuleNameOneOf("unclosed-expression"),
		},
		title:   staticTitle("Unclosed expression"),
		message: staticTitle("This **{{** has no matching **}}**."),
		block:   blockFromRuleName,
	}
}
