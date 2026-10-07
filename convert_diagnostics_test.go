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

func TestConvertDiagnosticParamUnitDropsPerRender(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ L arr = [1] }}\n\n{{ L arr = [] }}\n\n~ L arr = []\nItem: {{ arr[0] }}"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<p>Item: 1</p><p>Item: </p>", buf.String())
	want := []Diagnostic{{
		Code:    CodeArrayIndexOutOfRange,
		Message: "The index used for parameter **arr** is out of range.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 56, Line: 6, Column: 7},
			End:   Position{Offset: 68, Line: 6, Column: 19},
		},
		Calls: []Call{{
			Name:   "L",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 19, Line: 3, Column: 1},
				End:   Position{Offset: 35, Line: 3, Column: 17},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticBlockValueInlineDropsPerRender(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ W c = IN }}\n\n{{ W c = BL }}\n\n~ W c = NO_MATTER\nHello {{ c }}\n\n~ IN\ninline\n\n~ BL\n# block"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<p>Hello inline</p><p>Hello </p>", buf.String())
	want := []Diagnostic{{
		Code:    CodeInvalidComponentUsage,
		Message: "The component **BL** is a block component and cannot be used inline.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 56, Line: 6, Column: 7},
			End:   Position{Offset: 63, Line: 6, Column: 14},
		},
		Calls: []Call{{
			Name:   "W",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 16, Line: 3, Column: 1},
				End:   Position{Offset: 30, Line: 3, Column: 15},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticParamCompCallDropsPerRender(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ W c = A }}\n\n{{ W c = B }}\n\n~ W c = NO_MATTER\n{{ c x = \"hi\" }}\n\n~ A x = \"\"\nA {{ x }}\n\n~ B\nB"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<p>A hi</p>", buf.String())
	want := []Diagnostic{{
		Code:    CodeUnknownParameter,
		Message: "The parameter **x** is not defined for this component.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 48, Line: 6, Column: 1},
			End:   Position{Offset: 64, Line: 6, Column: 17},
		},
		Calls: []Call{{
			Name:   "W",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 15, Line: 3, Column: 1},
				End:   Position{Offset: 28, Line: 3, Column: 14},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticDuplicateArgumentDropsCallGivingIt(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ WRAP content = CARD(title = \"Bound\") }}\n\n~ WRAP content = NO_MATTER\n{{ content title = \"Caller\" }}\n\n~ CARD title = \"\"\n# {{ title }}"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "", buf.String())
	want := []Diagnostic{{
		Code:    CodeDuplicateArgument,
		Message: "The parameter **title** of component **CARD** is already bound.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 71, Line: 4, Column: 1},
			End:   Position{Offset: 101, Line: 4, Column: 31},
		},
		Calls: []Call{{
			Name:   "WRAP",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 0, Line: 1, Column: 1},
				End:   Position{Offset: 42, Line: 1, Column: 43},
			},
		}},
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

func TestConvertDroppedUnitRendersNothing(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ FOO }}\n\nHello {{ FOO }} world"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<p>Hello  world</p>", buf.String())
	assert.Len(t, diags, 2)
}

func TestConvertDiagnosticDropsOnlyUnitInLinkText(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("[Docs {{ label }}](/docs)"), &buf)
	require.NoError(t, err)
	assert.Equal(t, `<p><a href="/docs">Docs </a></p>`, buf.String())
	want := []Diagnostic{{
		Code:    CodeInvalidParameterUsage,
		Message: "Parameters cannot be used in the root context.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 6, Line: 1, Column: 7},
			End:   Position{Offset: 17, Line: 1, Column: 18},
		},
		Calls: nil,
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticDropsLinkForURLUnit(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ COMP }}\n\n~ COMP\n[{{ label }}]({{ target }})"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<p></p>", buf.String())
	want := []Diagnostic{{
		Code:    CodeUnknownParameter,
		Message: "The parameter **target** is not defined for this component.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 19, Line: 4, Column: 1},
			End:   Position{Offset: 46, Line: 4, Column: 28},
		},
		Calls: []Call{{
			Name:   "COMP",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 0, Line: 1, Column: 1},
				End:   Position{Offset: 10, Line: 1, Column: 11},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticForwardedArgTypeDropsPerRender(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ W a = [1] }}\n\n{{ W a = [\"x\"] }}\n\n~ W a = []\n{{ C n = a[0] }}\n\n~ C n = 0\nN {{ n }}"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<p>N 1</p>", buf.String())
	want := []Diagnostic{{
		Code:    CodeWrongArgumentType,
		Message: "The parameter **n** has the wrong type.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 47, Line: 6, Column: 1},
			End:   Position{Offset: 63, Line: 6, Column: 17},
		},
		Calls: []Call{{
			Name:   "W",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 17, Line: 3, Column: 1},
				End:   Position{Offset: 34, Line: 3, Column: 18},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticInfiniteCallPerEntry(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ A }}\n\n{{ B }}\n\n~ A\n# a\n{{ B }}\n\n~ B\n# b\n{{ A }}"), &buf)
	require.NoError(t, err)
	assert.Equal(t, "<h1>a</h1><h1>b</h1><h1>b</h1><h1>a</h1>", buf.String())
	want := []Diagnostic{
		{
			Code:    CodeInfiniteComponentCall,
			Message: "The call to component **A** creates an infinite loop and was skipped.",
			Source:  nil,
			Range: Range{
				Start: Position{Offset: 43, Line: 11, Column: 1},
				End:   Position{Offset: 50, Line: 11, Column: 8},
			},
			Calls: []Call{
				{
					Name:   "A",
					Kind:   FrameLocal,
					Source: nil,
					Range: Range{
						Start: Position{Offset: 0, Line: 1, Column: 1},
						End:   Position{Offset: 7, Line: 1, Column: 8},
					},
				},
				{
					Name:   "B",
					Kind:   FrameLocal,
					Source: nil,
					Range: Range{
						Start: Position{Offset: 26, Line: 7, Column: 1},
						End:   Position{Offset: 33, Line: 7, Column: 8},
					},
				},
			},
		},
		{
			Code:    CodeInfiniteComponentCall,
			Message: "The call to component **B** creates an infinite loop and was skipped.",
			Source:  nil,
			Range: Range{
				Start: Position{Offset: 26, Line: 7, Column: 1},
				End:   Position{Offset: 33, Line: 7, Column: 8},
			},
			Calls: []Call{
				{
					Name:   "B",
					Kind:   FrameLocal,
					Source: nil,
					Range: Range{
						Start: Position{Offset: 9, Line: 3, Column: 1},
						End:   Position{Offset: 16, Line: 3, Column: 8},
					},
				},
				{
					Name:   "A",
					Kind:   FrameLocal,
					Source: nil,
					Range: Range{
						Start: Position{Offset: 43, Line: 11, Column: 1},
						End:   Position{Offset: 50, Line: 11, Column: 8},
					},
				},
			},
		},
	}
	assert.Equal(t, want, diags)
}

func TestConvertDiagnosticImageDropsPerRender(t *testing.T) {
	var buf bytes.Buffer
	diags, err := New().Convert([]byte("{{ CARD m = \"image/jpeg\" }}\n\n{{ CARD m = \"image/svg+xml\" }}\n\n~ CARD m = \"\"\n{{ IMAGE media = { url: \"/a.jpg\", width: 1, height: 1, mime-type: m } alt = \"a\" }}"), &buf)
	require.NoError(t, err)
	assert.Equal(t, `<img src="/a.jpg" alt="a" width="1" height="1">`, buf.String())
	want := []Diagnostic{{
		Code:    CodeUnsupportedMimeType,
		Message: "The mime-type **image/svg+xml** is unsupported.",
		Source:  nil,
		Range: Range{
			Start: Position{Offset: 75, Line: 6, Column: 1},
			End:   Position{Offset: 157, Line: 6, Column: 83},
		},
		Calls: []Call{{
			Name:   "CARD",
			Kind:   FrameLocal,
			Source: nil,
			Range: Range{
				Start: Position{Offset: 29, Line: 3, Column: 1},
				End:   Position{Offset: 59, Line: 3, Column: 31},
			},
		}},
	}}
	assert.Equal(t, want, diags)
}
