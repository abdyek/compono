package compono

import (
	"github.com/umono-cms/compono/ast"
)

// DiagnosticCode is the fixed name of the kind of a diagnostic. It is the
// kebab-case form of the error title. Several situations may share a code; the
// message tells them apart. A code is never renamed.
type DiagnosticCode string

const (
	CodeArrayIndexOutOfRange       DiagnosticCode = "array-index-out-of-range"
	CodeDuplicateArgument          DiagnosticCode = "duplicate-argument"
	CodeDuplicateVariant           DiagnosticCode = "duplicate-variant"
	CodeInconsistentAspectRatio    DiagnosticCode = "inconsistent-aspect-ratio"
	CodeInfiniteComponentCall      DiagnosticCode = "infinite-component-call"
	CodeInvalidBuiltinArguments    DiagnosticCode = "invalid-built-in-arguments"
	CodeInvalidComponentUsage      DiagnosticCode = "invalid-component-usage"
	CodeInvalidContextAccess       DiagnosticCode = "invalid-context-access"
	CodeInvalidDimension           DiagnosticCode = "invalid-dimension"
	CodeInvalidParameterAccess     DiagnosticCode = "invalid-parameter-access"
	CodeInvalidParameterDefinition DiagnosticCode = "invalid-parameter-definition"
	CodeInvalidParameterUsage      DiagnosticCode = "invalid-parameter-usage"
	CodeMissingArgument            DiagnosticCode = "missing-argument"
	CodeNotComponentParameter      DiagnosticCode = "not-component-parameter"
	CodeUnknownComponent           DiagnosticCode = "unknown-component"
	CodeUnknownKey                 DiagnosticCode = "unknown-key"
	CodeUnknownParameter           DiagnosticCode = "unknown-parameter"
	CodeUnknownRecordKey           DiagnosticCode = "unknown-record-key"
	CodeUnsupportedMimeType        DiagnosticCode = "unsupported-mime-type"
	CodeWrongArgumentType          DiagnosticCode = "wrong-argument-type"
)

// Position is a point in a source: a 0-based byte offset, a 1-based line and a
// 1-based column counted in runes.
type Position = ast.Position

// Range is a [Start, End) range in a source.
type Range struct {
	Start Position
	End   Position
}

// Diagnostic is an error in a source. The unit the error belongs to is dropped
// from the output; the rest of the output is written and valid.
type Diagnostic struct {
	// Code is the kind of the error. Programs look at the code, not at the
	// message.
	Code DiagnosticCode
	// Message is an English description. Values are emphasized with **. It is
	// not HTML.
	Message string
	// Source is the source the dropped unit is written in. It is empty for the
	// converted source. For a global it is its scope path ending with its own
	// name: [LAYOUT CARD] for the sub component CARD of LAYOUT.
	Source []string
	// Range is the range of the dropped unit in Source.
	Range Range
	// Calls are the component calls around the dropped unit, outermost first.
	// It is empty when the unit is in the converted source itself.
	Calls []Call
}

// Call is a component call around a dropped unit.
type Call struct {
	Name string
	Kind FrameKind
	// Source is the source the call is written in, as in Diagnostic.
	Source []string
	// Range is the range of the call in Source.
	Range Range
}
