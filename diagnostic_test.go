package compono

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var diagnosticCodeTitles = []struct {
	constant DiagnosticCode
	title    string
}{
	{CodeArrayIndexOutOfRange, "Array index out of range"},
	{CodeDuplicateArgument, "Duplicate argument"},
	{CodeDuplicateVariant, "Duplicate variant"},
	{CodeInconsistentAspectRatio, "Inconsistent aspect ratio"},
	{CodeInfiniteComponentCall, "Infinite component call"},
	{CodeInvalidBuiltinArguments, "Invalid built-in arguments"},
	{CodeInvalidComponentUsage, "Invalid component usage"},
	{CodeInvalidContextAccess, "Invalid context access"},
	{CodeInvalidDimension, "Invalid dimension"},
	{CodeInvalidParameterAccess, "Invalid parameter access"},
	{CodeInvalidParameterDefinition, "Invalid parameter definition"},
	{CodeInvalidParameterUsage, "Invalid parameter usage"},
	{CodeMissingArgument, "Missing argument"},
	{CodeNotComponentParameter, "Not component parameter"},
	{CodeUnknownComponent, "Unknown component"},
	{CodeUnknownKey, "Unknown key"},
	{CodeUnknownParameter, "Unknown parameter"},
	{CodeUnknownRecordKey, "Unknown record key"},
	{CodeUnsupportedMimeType, "Unsupported mime-type"},
	{CodeWrongArgumentType, "Wrong argument type"},
}

func TestDiagnosticCodes(t *testing.T) {
	for _, tt := range diagnosticCodeTitles {
		t.Run(tt.title, func(t *testing.T) {
			expected := strings.ReplaceAll(strings.ToLower(tt.title), " ", "-")
			assert.Equal(t, expected, string(tt.constant))
		})
	}
}

func TestDiagnosticCodesAreUnique(t *testing.T) {
	seen := make(map[string]string, len(diagnosticCodeTitles))

	for _, tt := range diagnosticCodeTitles {
		previous, exists := seen[string(tt.constant)]
		assert.Falsef(t, exists, "duplicate code %q for %q and %q", tt.constant, previous, tt.title)
		seen[string(tt.constant)] = tt.title
	}
}
