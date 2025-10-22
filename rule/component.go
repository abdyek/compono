package rule

import "github.com/umono-cms/compono/selector"

// Global and local components definition wrapper
type compDefWrapper struct {
	scalable
}

func newCompDefWrapper() Rule {
	return &compDefWrapper{
		scalable: scalable{
			rules: []Rule{
				newLocalCompDef(),
			},
		},
	}
}

func (_ *compDefWrapper) Name() string {
	return "comp-def-wrapper"
}

func (_ *compDefWrapper) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewSinceFirstMatchInner(`\n*~\s+[A-Z0-9]+(?:_[A-Z0-9]+)*\s*\n`),
	}
}

func (cdw *compDefWrapper) Rules() []Rule {
	return cdw.rules
}

// Local component definition
type localCompDef struct {
	scalable
}

func newLocalCompDef() Rule {
	return &localCompDef{
		scalable: scalable{
			rules: []Rule{
				newLocalCompDefHead(),
				newLocalCompDefContent(),
			},
		},
	}
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

func (lcd *localCompDef) Rules() []Rule {
	return lcd.rules
}

// Local component definition head
type localCompDefHead struct {
	scalable
}

func newLocalCompDefHead() Rule {
	return &localCompDefHead{
		scalable: scalable{
			rules: []Rule{
				newLocalCompName(),
				newLocalCompParams(),
			},
		},
	}
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

func (lcdh *localCompDefHead) Rules() []Rule {
	return lcdh.rules
}

// Local component name
type localCompName struct {
	scalable
}

func newLocalCompName() Rule {
	return &localCompName{
		scalable: scalable{
			rules: []Rule{},
		},
	}
}

func (_ *localCompName) Name() string {
	return "local-comp-name"
}

func (_ *localCompName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\n~\s+`, `\s*\n`),
	}
}

func (lcn *localCompName) Rules() []Rule {
	return lcn.rules
}

// Local component parameters
type localCompParams struct {
	scalable
}

func newLocalCompParams() Rule {
	return &localCompParams{
		scalable: scalable{
			rules: []Rule{
				newLocalCompParam(),
			},
		},
	}
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

func (lcp *localCompParams) Rules() []Rule {
	return lcp.rules
}

// Local component parameter
type localCompParam struct {
	scalable
}

func newLocalCompParam() Rule {
	return &localCompParam{
		scalable: scalable{
			rules: []Rule{
				newLocalCompParamName(),
				newLocalCompParamType(),
			},
		},
	}
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

func (lcp *localCompParam) Rules() []Rule {
	return lcp.rules
}

// Local component parameter name
type localCompParamName struct {
	scalable
}

func newLocalCompParamName() Rule {
	return &localCompParamName{
		scalable: scalable{
			rules: []Rule{},
		},
	}
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

func (lcpn *localCompParamName) Rules() []Rule {
	return lcpn.rules
}

// Local component parameter type
type localCompParamType struct {
	scalable
}

func newLocalCompParamType() Rule {
	return &localCompParamType{
		scalable: scalable{
			rules: []Rule{
				newLocalCompStringParam(),
				newLocalCompNumberParam(),
				newLocalCompBoolParam(),
			},
		},
	}
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

func (lcpt *localCompParamType) Rules() []Rule {
	return lcpt.rules
}

// Local component's string parameter
type localCompStringParam struct {
	scalable
}

func newLocalCompStringParam() Rule {
	return &localCompStringParam{
		scalable: scalable{
			rules: []Rule{
				newLocalCompParamDefaValue(),
			},
		},
	}
}

func (_ *localCompStringParam) Name() string {
	return "local-comp-string-param"
}

func (_ *localCompStringParam) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*"`, `"[\s\n\r]*`),
	}
}

func (lcsp *localCompStringParam) Rules() []Rule {
	return lcsp.rules
}

// Local component's number parameter
type localCompNumberParam struct {
	scalable
}

func newLocalCompNumberParam() Rule {
	return &localCompNumberParam{
		scalable: scalable{
			rules: []Rule{
				newLocalCompParamDefaValue(),
			},
		},
	}
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

func (lcnp *localCompNumberParam) Rules() []Rule {
	return lcnp.rules
}

// Local component's bool parameter
type localCompBoolParam struct {
	scalable
}

func newLocalCompBoolParam() Rule {
	return &localCompBoolParam{
		scalable: scalable{
			rules: []Rule{
				newLocalCompParamDefaValue(),
			},
		},
	}
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

func (lcbp *localCompBoolParam) Rules() []Rule {
	return lcbp.rules
}

// Local component parameter default value
type localCompParamDefaValue struct {
	scalable
}

func newLocalCompParamDefaValue() Rule {
	return &localCompParamDefaValue{
		scalable: scalable{
			rules: []Rule{},
		},
	}
}

func (_ *localCompParamDefaValue) Name() string {
	return "local-comp-param-defa-value"
}

func (_ *localCompParamDefaValue) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (lcpdf *localCompParamDefaValue) Rules() []Rule {
	return lcpdf.rules
}

// Local component definition content
type localCompDefContent struct {
	scalable
}

func newLocalCompDefContent() Rule {
	return &localCompDefContent{
		scalable: scalable{
			rules: []Rule{
				newParamRef(),
			},
		},
	}
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

func (lcdc *localCompDefContent) Rules() []Rule {
	return lcdc.rules
}

// Parameter reference
type paramRef struct {
	scalable
}

func newParamRef() Rule {
	return &paramRef{
		scalable: scalable{
			rules: []Rule{
				newParamRefName(),
			},
		},
	}
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

func (pr *paramRef) Rules() []Rule {
	return pr.rules
}

// Parameter reference's name
type paramRefName struct {
	scalable
}

func newParamRefName() Rule {
	return &paramRefName{
		scalable: scalable{
			rules: []Rule{},
		},
	}
}

func (_ *paramRefName) Name() string {
	return "param-ref-name"
}

func (_ *paramRefName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\$`, `\z`),
	}
}

func (prn *paramRefName) Rules() []Rule {
	return prn.rules
}

// Component call
type compCall struct {
	scalable
}

func newCompCall() Rule {
	return &compCall{
		scalable: scalable{
			rules: []Rule{
				newCompCallName(),
				newCompCallArgs(),
			},
		},
	}
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

func (cc *compCall) Rules() []Rule {
	return cc.rules
}

// Component call name
type compCallName struct {
	scalable
}

func newCompCallName() Rule {
	return &compCallName{
		scalable: scalable{
			rules: []Rule{},
		},
	}
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

func (ccn *compCallName) Rules() []Rule {
	return ccn.rules
}

// Commponent call arguments
type compCallArgs struct {
	scalable
}

func newCompCallArgs() Rule {
	return &compCallArgs{
		scalable: scalable{
			rules: []Rule{
				newCompCallArg(),
			},
		},
	}
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

func (cca *compCallArgs) Rules() []Rule {
	return cca.rules
}

// Component call argument
type compCallArg struct {
	scalable
}

func newCompCallArg() Rule {
	return &compCallArg{
		scalable: scalable{
			rules: []Rule{
				newCompCallArgName(),
				newCompCallArgType(),
			},
		},
	}
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

func (cca *compCallArg) Rules() []Rule {
	return cca.rules
}

// Component call argument name
type compCallArgName struct {
	scalable
}

func newCompCallArgName() Rule {
	return &compCallArgName{
		scalable: scalable{
			rules: []Rule{},
		},
	}
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

func (ccan *compCallArgName) Rules() []Rule {
	return ccan.rules
}

// Component call argument type
type compCallArgType struct {
	scalable
}

func newCompCallArgType() Rule {
	return &compCallArgName{
		scalable: scalable{
			rules: []Rule{
				newCompCallStringArg(),
				newCompCallNumberArg(),
				newCompCallBoolArg(),
			},
		},
	}
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

func (ccat *compCallArgType) Rules() []Rule {
	return ccat.rules
}

// Component call's string argument
type compCallStringArg struct {
	scalable
}

func newCompCallStringArg() Rule {
	return &compCallStringArg{
		scalable: scalable{
			rules: []Rule{
				newCompCallArgValue(),
			},
		},
	}
}

func (_ *compCallStringArg) Name() string {
	return "comp-call-string-arg"
}

func (_ *compCallStringArg) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`[\s\n\r]*"`, `"[\s\n\r]*`),
	}
}

func (ccsa *compCallStringArg) Rules() []Rule {
	return ccsa.rules
}

// Component call's number argument
type compCallNumberArg struct {
	scalable
}

func newCompCallNumberArg() Rule {
	return &compCallNumberArg{
		scalable: scalable{
			rules: []Rule{
				newCompCallArgValue(),
			},
		},
	}
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

func (ccna *compCallNumberArg) Rules() []Rule {
	return ccna.rules
}

// Component call's bool argument
type compCallBoolArg struct {
	scalable
}

func newCompCallBoolArg() Rule {
	return &compCallBoolArg{
		scalable: scalable{
			rules: []Rule{
				newCompCallArgValue(),
			},
		},
	}
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

func (ccba *compCallBoolArg) Rules() []Rule {
	return ccba.rules
}

// Component call argument value
type compCallArgValue struct {
	scalable
}

func newCompCallArgValue() Rule {
	return &compCallArgValue{
		scalable: scalable{
			rules: []Rule{},
		},
	}
}

func (_ *compCallArgValue) Name() string {
	return "comp-call-arg-value"
}

func (_ *compCallArgValue) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewAll(),
	}
}

func (ccav *compCallArgValue) Rules() []Rule {
	return ccav.rules
}
