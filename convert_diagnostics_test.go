package compono

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertReturnsNoDiagnosticsWithoutErrors(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("# Hi"), &buf)
	require.NoError(t, err)
	assert.Empty(t, diags)
	assert.Equal(t, "<h1>Hi</h1>", buf.String())
}

func TestConvertDiagnosticInConvertedSource(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("Hello {{ FOO }} world"), &buf)
	require.NoError(t, err)
	want := []Diagnostic{{
		Code:    CodeUnknownComponent,
		Message: "The component **FOO** is not defined or not registered.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 6, Line: 1, Column: 7},
			End:   Position{Offset: 15, Line: 1, Column: 16},
		},
		Calls: nil,
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticColumnCountsRunes(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("ç\n\nA {{ FOO }}"), &buf)
	require.NoError(t, err)
	want := []Diagnostic{{
		Code:    CodeUnknownComponent,
		Message: "The component **FOO** is not defined or not registered.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 6, Line: 3, Column: 3},
			End:   Position{Offset: 15, Line: 3, Column: 12},
		},
		Calls: nil,
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticsInOutputOrder(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ A }}\n\nText {{ B }}"), &buf)
	require.NoError(t, err)
	want := []Diagnostic{
		{
			Code:    CodeUnknownComponent,
			Message: "The component **A** is not defined or not registered.",
			Source:  nil,
			Range: Range{
				Start: Position{Offset: 0, Line: 1, Column: 1},
				End:   Position{Offset: 7, Line: 1, Column: 8},
			},
			Calls: nil,
		},
		{
			Code:    CodeUnknownComponent,
			Message: "The component **B** is not defined or not registered.",
			Source:  nil,
			Range: Range{
				Start: Position{Offset: 14, Line: 3, Column: 6},
				End:   Position{Offset: 21, Line: 3, Column: 13},
			},
			Calls: nil,
		},
	}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticPerCallOfGlobal(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ G }}\n\n{{ G }}"), &buf,
		WithGlobalComponent("G", []byte("Hi {{ FOO }}")),
	)
	require.NoError(t, err)
	want := []Diagnostic{
		{
			Code:    CodeUnknownComponent,
			Message: "The component **FOO** is not defined or not registered.",
			Source:  []string{"G"},
			Range: Range{
				Start: Position{Offset: 3, Line: 1, Column: 4},
				End:   Position{Offset: 12, Line: 1, Column: 13},
			},
			Calls: []Call{{
				Name:   "G",
				Kind:   FrameGlobal,
				Source: nil,
				Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 1},
					End:   Position{Offset: 7, Line: 1, Column: 8},
				},
			}},
		},
		{
			Code:    CodeUnknownComponent,
			Message: "The component **FOO** is not defined or not registered.",
			Source:  []string{"G"},
			Range: Range{
				Start: Position{Offset: 3, Line: 1, Column: 4},
				End:   Position{Offset: 12, Line: 1, Column: 13},
			},
			Calls: []Call{{
				Name:   "G",
				Kind:   FrameGlobal,
				Source: nil,
				Range: Range{
					Start: Position{Offset: 9, Line: 3, Column: 1},
					End:   Position{Offset: 16, Line: 3, Column: 8},
				},
			}},
		},
	}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticInLocalComponent(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ L }}\n\n~ L\n{{ FOO }}"), &buf)
	require.NoError(t, err)
	want := []Diagnostic{{
		Code:    CodeUnknownComponent,
		Message: "The component **FOO** is not defined or not registered.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 13, Line: 4, Column: 1},
			End:   Position{Offset: 22, Line: 4, Column: 10},
		},
		Calls: []Call{{
			Name:   "L",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 0, Line: 1, Column: 1},
				End:   Position{Offset: 7, Line: 1, Column: 8},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticInSubComponent(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ LAYOUT }}"), &buf,
		WithGlobalComponent("LAYOUT", []byte("{{ CARD }}"),
			WithGlobalComponent("CARD", []byte("{{ FOO }}")),
		),
	)
	require.NoError(t, err)
	want := []Diagnostic{{
		Code:    CodeUnknownComponent,
		Message: "The component **FOO** is not defined or not registered.",
		Source:  []string{"LAYOUT", "CARD"},
		Range: Range{
			Start: Position{Offset: 0, Line: 1, Column: 1},
			End:   Position{Offset: 9, Line: 1, Column: 10},
		},
		Calls: []Call{
			{
				Name:   "LAYOUT",
				Kind:   FrameGlobal,
				Source: nil,
				Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 1},
					End:   Position{Offset: 12, Line: 1, Column: 13},
				},
			},
			{
				Name:   "CARD",
				Kind:   FrameGlobal,
				Source: []string{"LAYOUT"},
				Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 1},
					End:   Position{Offset: 10, Line: 1, Column: 11},
				},
			},
		},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticThroughComponentParameter(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ L c = G }}\n\n~ L c = NO_MATTER\n{{ c }}"), &buf,
		WithGlobalComponent("G", []byte("{{ FOO }}")),
	)
	require.NoError(t, err)
	want := []Diagnostic{{
		Code:    CodeUnknownComponent,
		Message: "The component **FOO** is not defined or not registered.",
		Source:  []string{"G"},
		Range: Range{
			Start: Position{Offset: 0, Line: 1, Column: 1},
			End:   Position{Offset: 9, Line: 1, Column: 10},
		},
		Calls: []Call{
			{
				Name:   "L",
				Kind:   FrameLocal,
				Source: nil,
				Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 1},
					End:   Position{Offset: 13, Line: 1, Column: 14},
				},
			},
			{
				Name:   "G",
				Kind:   FrameGlobal,
				Source: nil,
				Range: Range{
					Start: Position{Offset: 33, Line: 4, Column: 1},
					End:   Position{Offset: 40, Line: 4, Column: 8},
				},
			},
		},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticOnceForBlockErrorInParagraph(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("A\n{{ FOO }}\nB"), &buf)
	require.NoError(t, err)
	want := []Diagnostic{{
		Code:    CodeUnknownComponent,
		Message: "The component **FOO** is not defined or not registered.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 2, Line: 2, Column: 1},
			End:   Position{Offset: 11, Line: 2, Column: 10},
		},
		Calls: nil,
	}}
	assert.Equal(t, want, diags)
}

func TestConvertFatalErrorReturnsNoDiagnostics(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ FOO }}"), &buf, WithIsolatedScope())
	assert.Error(t, err)
	assert.Nil(t, diags)
	assert.Empty(t, buf.String())
}

func TestConvertEmptySourceReturnsNoDiagnostics(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte(""), &buf)
	assert.NoError(t, err)
	assert.Nil(t, diags)
}
