package rule

import "github.com/umono-cms/compono/selector"

// Local components definition wrapper
type localCompDefWrapper struct{}

func newLocalCompDefWrapper() Rule {
	return &localCompDefWrapper{}
}

func (_ *localCompDefWrapper) Name() string {
	return "local-comp-def-wrapper"
}

func (_ *localCompDefWrapper) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewSinceFirstMatchInner(`\n*~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*\s*\n`),
	}
}

func (_ *localCompDefWrapper) Rules() []Rule {
	return []Rule{
		newLocalCompDef(),
	}
}

// Local component definition
type localCompDef struct{}

func newLocalCompDef() Rule {
	return &localCompDef{}
}

func (_ *localCompDef) Name() string {
	return "local-comp-def"
}

func (_ *localCompDef) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*\s*\n`, `\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*\s*\n|\z`)
	return []selector.Selector{
		seli,
	}
}

func (_ *localCompDef) Rules() []Rule {
	return []Rule{
		newLocalCompDefHead(),
		newLocalCompDefContent(),
	}
}

// Local component definition head
type localCompDefHead struct{}

func newLocalCompDefHead() Rule {
	return &localCompDefHead{}
}

func (_ *localCompDefHead) Name() string {
	return "local-comp-def-head"
}

func (_ *localCompDefHead) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`\n~\s+`, `\s*\n\n|\z`)
	return []selector.Selector{
		se,
	}
}

func (_ *localCompDefHead) Rules() []Rule {
	return []Rule{
		newLocalCompName(),
		newLocalCompParams(),
	}
}

// Local component name
type localCompName struct{}

func newLocalCompName() Rule {
	return &localCompName{}
}

func (_ *localCompName) Name() string {
	return "local-comp-name"
}

func (_ *localCompName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\n~\s+`, `\s*\n`),
	}
}

func (_ *localCompName) Rules() []Rule {
	return []Rule{}
}

// Local component parameters
type localCompParams struct{}

func newLocalCompParams() Rule {
	return &localCompParams{}
}

func (_ *localCompParams) Name() string {
	return "local-comp-params"
}

func (_ *localCompParams) Selectors() []selector.Selector {
	se, _ := selector.NewStartEnd(`\n`, `.`)
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false)`)
	return []selector.Selector{
		selector.NewBounds(se, p),
	}
}

func (_ *localCompParams) Rules() []Rule {
	return []Rule{
		newLocalCompParam(),
	}
}

// Local component parameter
type localCompParam struct{}

func newLocalCompParam() Rule {
	return &localCompParam{}
}

func (_ *localCompParam) Name() string {
	return "local-comp-param"
}

func (_ *localCompParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false)`)
	return []selector.Selector{
		p,
	}
}

func (_ *localCompParam) Rules() []Rule {
	return []Rule{
		newLocalCompParamName(),
		newLocalCompParamType(),
	}
}

// Local component parameter name
type localCompParamName struct{}

func newLocalCompParamName() Rule {
	return &localCompParamName{}
}

func (_ *localCompParamName) Name() string {
	return "local-comp-param-name"
}

func (_ *localCompParamName) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`([a-z][a-z0-9-]*)\s*`, `=`)
	return []selector.Selector{
		seli,
	}
}

func (_ *localCompParamName) Rules() []Rule {
	return []Rule{}
}

// Local component parameter type
type localCompParamType struct{}

func newLocalCompParamType() Rule {
	return &localCompParamType{}
}

func (_ *localCompParamType) Name() string {
	return "local-comp-param-type"
}

func (_ *localCompParamType) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false)`)
	return []selector.Selector{
		p,
	}
}

func (_ *localCompParamType) Rules() []Rule {
	return []Rule{
		newLocalCompStringParam(),
		newLocalCompNumberParam(),
		newLocalCompBoolParam(),
	}
}

// Local component's string parameter
type localCompStringParam struct{}

func newLocalCompStringParam() Rule {
	return &localCompStringParam{}
}

func (_ *localCompStringParam) Name() string {
	return "local-comp-string-param"
}

func (_ *localCompStringParam) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*"`, `"[\s\n\r]*`),
	}
}

func (_ *localCompStringParam) Rules() []Rule {
	return []Rule{
		newLocalCompParamDefaValue(),
	}
}

// Local component's number parameter
type localCompNumberParam struct{}

func newLocalCompNumberParam() Rule {
	return &localCompNumberParam{}
}

func (_ *localCompNumberParam) Name() string {
	return "local-comp-number-param"
}

func (_ *localCompNumberParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\d+(?:\.\d+)?`)
	return []selector.Selector{
		p,
	}
}

func (_ *localCompNumberParam) Rules() []Rule {
	return []Rule{
		newLocalCompParamDefaValue(),
	}
}

// Local component's bool parameter
type localCompBoolParam struct{}

func newLocalCompBoolParam() Rule {
	return &localCompBoolParam{}
}

func (_ *localCompBoolParam) Name() string {
	return "local-comp-bool-param"
}

func (_ *localCompBoolParam) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`true|false`)
	return []selector.Selector{
		p,
	}
}

func (_ *localCompBoolParam) Rules() []Rule {
	return []Rule{
		newLocalCompParamDefaValue(),
	}
}

// Local component parameter default value
type localCompParamDefaValue struct{}

func newLocalCompParamDefaValue() Rule {
	return &localCompParamDefaValue{}
}

func (_ *localCompParamDefaValue) Name() string {
	return "local-comp-param-defa-value"
}

func (_ *localCompParamDefaValue) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *localCompParamDefaValue) Rules() []Rule {
	return []Rule{}
}

// Local component definition content
type localCompDefContent struct{}

func newLocalCompDefContent() Rule {
	return &localCompDefContent{}
}

func (_ *localCompDefContent) Name() string {
	return "local-comp-def-content"
}

func (_ *localCompDefContent) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`^`, `\n~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*|\z`)
	return []selector.Selector{
		seli,
	}
}

func (_ *localCompDefContent) Rules() []Rule {
	return []Rule{
		newParamRef(),
	}
}

// Parameter reference
type paramRef struct{}

func newParamRef() Rule {
	return &paramRef{}
}

func (_ *paramRef) Name() string {
	return "param-ref"
}

func (_ *paramRef) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\$[a-z][a-z0-9-]*`)
	return []selector.Selector{
		p,
	}
}

func (_ *paramRef) Rules() []Rule {
	return []Rule{
		newParamRefName(),
	}
}

// Parameter reference's name
type paramRefName struct{}

func newParamRefName() Rule {
	return &paramRefName{}
}

func (_ *paramRefName) Name() string {
	return "param-ref-name"
}

func (_ *paramRefName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\$`, `\z`),
	}
}

func (_ *paramRefName) Rules() []Rule {
	return []Rule{}
}

// Block component call
type blockCompCall struct {
	*compCall
}

func newBlockCompCall() Rule {
	cc := newCompCall()
	return &blockCompCall{
		compCall: cc.(*compCall),
	}
}

func (_ *blockCompCall) Name() string {
	return "block-comp-call"
}

// Inline component call
type inlineCompCall struct {
	*compCall
}

func newInlineCompCall() Rule {
	cc := newCompCall()
	return &inlineCompCall{
		compCall: cc.(*compCall),
	}
}

func (_ *inlineCompCall) Name() string {
	return "inline-comp-call"
}

// Component call
type compCall struct{}

func newCompCall() Rule {
	return &compCall{}
}

func (_ *compCall) Name() string {
	return "comp-call"
}

func (_ *compCall) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\{\{\s*[A-Z0-9]+(?:_[A-Z0-9]+)`, `\s*\}\}`)
	return []selector.Selector{
		seSelector,
	}
}

func (_ *compCall) Rules() []Rule {
	return []Rule{
		newCompCallName(),
		newCompCallArgs(),
	}
}

// Component call name
type compCallName struct{}

func newCompCallName() Rule {
	return &compCallName{}
}

func (_ *compCallName) Name() string {
	return "comp-call-name"
}

func (_ *compCallName) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\s*[A-Z0-9]+(?:_[A-Z0-9]+)*\s*`)
	return []selector.Selector{
		selector.NewFilter(p, func(source []byte, index [][2]int) [][2]int {
			if len(index) > 1 {
				return [][2]int{index[0]}
			}
			return [][2]int{}
		}),
	}
}

func (_ *compCallName) Rules() []Rule {
	return []Rule{}
}

// Commponent call arguments
type compCallArgs struct{}

func newCompCallArgs() Rule {
	return &compCallArgs{}
}

func (_ *compCallArgs) Name() string {
	return "comp-call-args"
}

func (_ *compCallArgs) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false)`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallArgs) Rules() []Rule {
	return []Rule{
		newCompCallArg(),
	}
}

// Component call argument
type compCallArg struct{}

func newCompCallArg() Rule {
	return &compCallArg{}
}

func (_ *compCallArg) Name() string {
	return "comp-call-arg"
}

func (_ *compCallArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`([a-z][a-z0-9-]*)[\s\n\r]*=[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false)`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallArg) Rules() []Rule {
	return []Rule{
		newCompCallArgName(),
		newCompCallArgType(),
	}
}

// Component call argument name
type compCallArgName struct{}

func newCompCallArgName() Rule {
	return &compCallArgName{}
}

func (_ *compCallArgName) Name() string {
	return "comp-call-arg-name"
}

func (_ *compCallArgName) Selectors() []selector.Selector {
	seli, _ := selector.NewStartEndLeftInner(`([a-z][a-z0-9-]*)\s*`, `=`)
	return []selector.Selector{
		seli,
	}
}

func (_ *compCallArgName) Rules() []Rule {
	return []Rule{}
}

// Component call argument type
type compCallArgType struct{}

func newCompCallArgType() Rule {
	return &compCallArgName{}
}

func (_ *compCallArgType) Name() string {
	return "comp-call-arg-type"
}

func (_ *compCallArgType) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`[\s\n\r]*(".*?"|\d+(?:\.\d+)?|true|false)`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallArgType) Rules() []Rule {
	return []Rule{
		newCompCallStringArg(),
		newCompCallNumberArg(),
		newCompCallBoolArg(),
	}
}

// Component call's string argument
type compCallStringArg struct{}

func newCompCallStringArg() Rule {
	return &compCallStringArg{}
}

func (_ *compCallStringArg) Name() string {
	return "comp-call-string-arg"
}

func (_ *compCallStringArg) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*"`, `"[\s\n\r]*`),
	}
}

func (_ *compCallStringArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Component call's number argument
type compCallNumberArg struct{}

func newCompCallNumberArg() Rule {
	return &compCallNumberArg{}
}

func (_ *compCallNumberArg) Name() string {
	return "comp-call-number-arg"
}

func (_ *compCallNumberArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`\d+(?:\.\d+)?`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallNumberArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Component call's bool argument
type compCallBoolArg struct{}

func newCompCallBoolArg() Rule {
	return &compCallBoolArg{}
}

func (_ *compCallBoolArg) Name() string {
	return "comp-call-arg-value"
}

func (_ *compCallBoolArg) Selectors() []selector.Selector {
	p, _ := selector.NewPattern(`true|false`)
	return []selector.Selector{
		p,
	}
}

func (_ *compCallBoolArg) Rules() []Rule {
	return []Rule{
		newCompCallArgValue(),
	}
}

// Component call argument value
type compCallArgValue struct{}

func newCompCallArgValue() Rule {
	return &compCallArgValue{}
}

func (_ *compCallArgValue) Name() string {
	return "comp-call-arg-value"
}

func (_ *compCallArgValue) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (_ *compCallArgValue) Rules() []Rule {
	return []Rule{}
}
