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
				newLocalCompDefName(),
				newLocalCompParams(),
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

// Local component definition name
type localCompDefName struct {
	scalable
}

func newLocalCompDefName() Rule {
	return &localCompDefName{
		scalable: scalable{
			rules: []Rule{},
		},
	}
}

func (_ *localCompDefName) Name() string {
	return "local-comp-def-name"
}

func (_ *localCompDefName) Selectors() []selector.Selector {
	return []selector.Selector{
		selector.NewStartEndInner(`\n~\s+`, `\s*\n`),
	}
}

func (lcdn *localCompDefName) Rules() []Rule {
	return lcdn.rules
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
	return []selector.Selector{
		// TODO: complete it
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
				newLocalCompParamDefaValue(),
			},
		},
	}
}

func (_ *localCompParam) Name() string {
	return "local-comp-param"
}

func (_ *localCompParam) Selectors() []selector.Selector {
	return []selector.Selector{
		// TODO: complete it
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
	return &localCompParam{
		scalable: scalable{
			rules: []Rule{},
		},
	}
}

func (_ *localCompParamName) Name() string {
	return "local-comp-param-name"
}

func (_ *localCompParamName) Selectors() []selector.Selector {
	return []selector.Selector{
		// TODO: complete it
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
			rules: []Rule{},
		},
	}
}

func (_ *localCompParamType) Name() string {
	return "local-comp-param-type"
}

func (_ *localCompParamType) Selectors() []selector.Selector {
	return []selector.Selector{
		// TODO: complete it
	}
}

func (lcpt *localCompParamType) Rules() []Rule {
	return lcpt.rules
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
		// TODO: complete it
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
			// NOTE: It must be BLOCK or INLINE
			rules: []Rule{},
		},
	}
}

func (_ *localCompDefContent) Name() string {
	return "local-comp-def-content"
}

func (_ *localCompDefContent) Selectors() []selector.Selector {
	return []selector.Selector{
		// TODO: complete it
	}
}

func (lcdc *localCompDefContent) Rules() []Rule {
	return lcdc.rules
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
			},
		},
	}
}

func (_ *compCall) Name() string {
	return "comp-call"
}

func (_ *compCall) Selectors() []selector.Selector {
	seSelector, _ := selector.NewStartEnd(`\{\{\s*`, `\s*\}\}`)
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
	return []selector.Selector{
		// TODO: complete it
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
	return []selector.Selector{
		// TODO: complete it
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
				newCompCallArgValue(),
			},
		},
	}
}

func (_ *compCallArg) Name() string {
	return "comp-call-arg"
}

func (_ *compCallArg) Selectors() []selector.Selector {
	return []selector.Selector{
		// TODO: complete it
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
	return []selector.Selector{
		// TODO: complete it
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
			rules: []Rule{},
		},
	}
}

func (_ *compCallArgType) Name() string {
	return "comp-call-arg-type"
}

func (_ *compCallArgType) Selectors() []selector.Selector {
	return []selector.Selector{
		// TODO: complete it
	}
}

func (ccat *compCallArgType) Rules() []Rule {
	return ccat.rules
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
		// TODO: complete it
	}
}

func (ccav *compCallArgValue) Rules() []Rule {
	return ccav.rules
}
