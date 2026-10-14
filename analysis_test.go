package compono

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umono-cms/compono/ast"
)

// span returns the range of the n-th (0-based) occurrence of text in source.
func span(t *testing.T, source, text string, n int) Range {
	t.Helper()
	idx := -1
	for i := 0; i <= n; i++ {
		next := strings.Index(source[idx+1:], text)
		require.NotEqual(t, -1, next, "text %q not found", text)
		idx += 1 + next
	}
	return Range{
		Start: ast.PositionAt([]byte(source), idx),
		End:   ast.PositionAt([]byte(source), idx+len(text)),
	}
}

func TestBuiltins(t *testing.T) {
	want := []string{"LINK", "IMAGE", "HEADER", "MAIN", "FOOTER", "ARTICLE", "ASIDE", "NAV", "DETAILS", "LIST"}
	got := Builtins()
	assert.Equal(t, want, got)
	got[0] = "CHANGED"
	assert.Equal(t, want, Builtins())
}

func TestMarkdownElements(t *testing.T) {
	want := []string{"blockquote", "br", "code-block", "em", "heading", "hr", "inline-code", "link", "list", "p", "strong", "table"}
	got := MarkdownElements()
	assert.Equal(t, want, got)
	got[0] = "changed"
	assert.Equal(t, want, MarkdownElements())
}

func TestAnalyzeEmptySource(t *testing.T) {
	analysis, err := Analyze(nil)
	require.NoError(t, err)
	assert.Equal(t, Analysis{}, analysis)
}

func TestAnalyzeIsDeterministic(t *testing.T) {
	source := []byte("{{ A x = B }}\n\n~ A x = NO_MATTER\n{{ x }} {{ Bad }}\n")
	first, err := Analyze(source)
	require.NoError(t, err)
	second, err := Analyze(source)
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestAnalyzeGlobal(t *testing.T) {
	source := "~ title! = \"\" c = D(a = 1) l = [Z]\n" +
		"// yorum\n" +
		"# {{ title }}\n" +
		"\n" +
		"{{ CARD x = 1 }} metin\n" +
		"\n" +
		"~ CARD x = 0 k = NO_MATTER\n" +
		"**{{ x }}** {{ LINK text = \"a\" url = \"/\" }}\n"

	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)
	cardDef := strings.TrimSuffix(source[strings.Index(source, "~ CARD"):], "\n")
	cardBody := strings.TrimSuffix(source[strings.Index(source, "**{{ x }}**"):], "\n")

	assert.Equal(t, []Param{
		{Name: "title", Type: "string", Required: true, Range: span(t, source, "title! = \"\"", 0)},
		{Name: "c", Type: "component", Range: span(t, source, "c = D(a = 1)", 0)},
		{Name: "l", Type: "array", Range: span(t, source, "l = [Z]", 0)},
	}, analysis.Params)

	assert.Equal(t, []Local{
		{
			Name: "CARD",
			Params: []Param{
				{Name: "x", Type: "integer", Range: span(t, source, "x = 0", 0)},
				{Name: "k", Type: "component", Range: span(t, source, "k = NO_MATTER", 0)},
			},
			Range: span(t, source, cardDef, 0),
		},
	}, analysis.Locals)

	assert.Equal(t, []Ref{
		{Name: "D", Kind: "default", Args: []string{"a"}, Range: span(t, source, "D(a = 1)", 0)},
		{Name: "Z", Kind: "default", Range: span(t, source, "Z", 0)},
		{Name: "CARD", Kind: "call", Local: true, Args: []string{"x"}, Range: span(t, source, "{{ CARD x = 1 }}", 0)},
		{Name: "NO_MATTER", Kind: "default", In: "CARD", Range: span(t, source, "NO_MATTER", 0)},
		{Name: "LINK", Kind: "call", In: "CARD", Args: []string{"text", "url"}, Range: span(t, source, "{{ LINK text = \"a\" url = \"/\" }}", 0)},
	}, analysis.Refs)

	assert.Equal(t, []Element{
		{Name: "heading", Range: span(t, source, "# {{ title }}", 0)},
		{Name: "p", Range: span(t, source, "{{ CARD x = 1 }} metin", 0)},
		{Name: "p", In: "CARD", Range: span(t, source, cardBody, 0)},
		{Name: "strong", In: "CARD", Range: span(t, source, "**{{ x }}**", 0)},
	}, analysis.Elements)

	assert.Equal(t, []Unit{
		{Kind: "markdown", Ref: -1, Range: span(t, source, "# {{ title }}", 0)},
		{Kind: "call", Ref: 2, Range: span(t, source, "{{ CARD x = 1 }}", 0)},
		{Kind: "text", Ref: -1, Range: span(t, source, "metin", 0)},
	}, analysis.Root)

	assert.Empty(t, analysis.Diagnostics)
}

func TestAnalyzePage(t *testing.T) {
	source := "{{ LAYOUT content = BODY }}\n" +
		"\n" +
		"~ BODY\n" +
		"Merhaba {{ IF x }}\n" +
		"\n" +
		"~ UNUSED\n" +
		"{{ Bad }}\n"

	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)

	assert.Nil(t, analysis.Params)
	assert.Equal(t, []Local{
		{Name: "BODY", Range: span(t, source, "~ BODY\nMerhaba {{ IF x }}", 0)},
		{Name: "UNUSED", Range: span(t, source, "~ UNUSED\n{{ Bad }}", 0)},
	}, analysis.Locals)
	assert.Equal(t, []Ref{
		{Name: "LAYOUT", Kind: "call", Args: []string{"content"}, Block: true, Range: span(t, source, "{{ LAYOUT content = BODY }}", 0)},
		{Name: "BODY", Kind: "value", Local: true, Range: span(t, source, "BODY", 0)},
	}, analysis.Refs)
	assert.Equal(t, []Element{
		{Name: "p", In: "BODY", Range: span(t, source, "Merhaba {{ IF x }}", 0)},
	}, analysis.Elements)
	assert.Equal(t, []Unit{
		{Kind: "call", Ref: 0, Range: span(t, source, "{{ LAYOUT content = BODY }}", 0)},
	}, analysis.Root)

	message := "The content between **{{** and **}}** must be a component call, a parameter reference or a context reference."
	assert.Equal(t, []Diagnostic{
		{Code: CodeInvalidExpression, Message: message, Range: span(t, source, "{{ IF x }}", 0)},
		{Code: CodeInvalidExpression, Message: message, Range: span(t, source, "{{ Bad }}", 0)},
	}, analysis.Diagnostics)
}

func TestAnalyzeRootUnits(t *testing.T) {
	source := "// yorum\n" +
		"{{ A }}\n" +
		"metin **k**\n" +
		"{{ B }}\n" +
		"\n" +
		"# Başlık\n"

	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)

	assert.Equal(t, []Ref{
		{Name: "A", Kind: "call", Block: true, Range: span(t, source, "{{ A }}", 0)},
		{Name: "B", Kind: "call", Block: true, Range: span(t, source, "{{ B }}", 0)},
	}, analysis.Refs)
	assert.Equal(t, []Unit{
		{Kind: "call", Ref: 0, Range: span(t, source, "{{ A }}", 0)},
		{Kind: "text", Ref: -1, Range: span(t, source, "metin", 0)},
		{Kind: "markdown", Ref: -1, Range: span(t, source, "**k**", 0)},
		{Kind: "call", Ref: 1, Range: span(t, source, "{{ B }}", 0)},
		{Kind: "markdown", Ref: -1, Range: span(t, source, "# Başlık", 0)},
	}, analysis.Root)
	assert.Empty(t, analysis.Diagnostics)
}

func TestAnalyzeElements(t *testing.T) {
	source := "a *e* `c` [l](/u)\n" +
		"b\n" +
		"\n" +
		"- x\n" +
		"\n" +
		"> q\n" +
		"\n" +
		"---\n" +
		"\n" +
		"| a |\n" +
		"| --- |\n" +
		"| 1 |\n" +
		"\n" +
		"```\n" +
		"kod\n" +
		"```\n"

	analysis, err := Analyze([]byte(source))
	require.NoError(t, err)

	br := Range{
		Start: Position{Offset: 17, Line: 1, Column: 18},
		End:   Position{Offset: 18, Line: 2, Column: 1},
	}
	assert.Equal(t, []Element{
		{Name: "p", Range: span(t, source, "a *e* `c` [l](/u)\nb", 0)},
		{Name: "em", Range: span(t, source, "*e*", 0)},
		{Name: "inline-code", Range: span(t, source, "`c`", 0)},
		{Name: "link", Range: span(t, source, "[l](/u)", 0)},
		{Name: "br", Range: br},
		{Name: "list", Range: span(t, source, "- x", 0)},
		{Name: "blockquote", Range: span(t, source, "> q", 0)},
		{Name: "p", Range: span(t, source, "q", 0)},
		{Name: "hr", Range: span(t, source, "---", 0)},
		{Name: "table", Range: span(t, source, "| a |\n| --- |\n| 1 |", 0)},
		{Name: "code-block", Range: span(t, source, "```\nkod\n```", 0)},
	}, analysis.Elements)
}
