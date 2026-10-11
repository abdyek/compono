package compono

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const urlReporterImageA = `{{ IMAGE media = { url: "/a.png", width: 1, height: 1, mime-type: "image/png" } alt = "" }}`

const urlReporterImageP = `{{ IMAGE media = { url: "/p.png", width: 1, height: 1, mime-type: "image/png" } alt = "" }}`

func convertWithReports(t *testing.T, source string, opts ...ConvertOption) (string, []URLReport) {
	t.Helper()
	var reports []URLReport
	opts = append(opts, WithURLReporter(func(r URLReport) {
		reports = append(reports, r)
	}))
	var buf bytes.Buffer
	_, err := New().Convert([]byte(source), &buf, opts...)
	require.NoError(t, err)
	return buf.String(), reports
}

func TestURLReporterReportsImageURLsInOutputOrder(t *testing.T) {
	source := `{{ IMAGE media = { url: "/main.jpg", width: 100, height: 50, mime-type: "image/jpeg", variants: [{ url: "/v-80.webp", width: 80, height: 40, mime-type: "image/webp" }, { url: "/v-40.webp", width: 40, height: 20, mime-type: "image/webp" }, { url: "/v-40.jpg", width: 40, height: 20, mime-type: "image/jpeg" }] } alt = "a" }}

[link](https://example.com) {{ LINK text = "l" url = "/l" }}`

	_, reports := convertWithReports(t, source)

	expected := []URLReport{
		{Element: "source", Attribute: "srcset", URL: "/v-40.webp"},
		{Element: "source", Attribute: "srcset", URL: "/v-80.webp"},
		{Element: "source", Attribute: "srcset", URL: "/v-40.jpg"},
		{Element: "img", Attribute: "src", URL: "/main.jpg"},
	}
	assert.Equal(t, expected, reports)
}

func TestURLReporterReportsNormalizedUnescapedURL(t *testing.T) {
	source := `{{ IMAGE media = { url: "  /a.png?x=1&y=2  ", width: 1, height: 1, mime-type: "image/png" } alt = "" }}`

	_, reports := convertWithReports(t, source)

	expected := []URLReport{
		{Element: "img", Attribute: "src", URL: "/a.png?x=1&y=2"},
	}
	assert.Equal(t, expected, reports)
}

func TestURLReporterSkipsDroppedImage(t *testing.T) {
	source := `{{ IMAGE media = { url: "data:x", width: 1, height: 1, mime-type: "image/png" } alt = "" }}
{{ IMAGE media = { url: "/ok.png", width: 1, height: 1, mime-type: "image/png" } alt = "" }}`

	_, reports := convertWithReports(t, source)

	expected := []URLReport{
		{Element: "img", Attribute: "src", URL: "/ok.png"},
	}
	assert.Equal(t, expected, reports)
}

func TestURLReporterReportsEveryOccurrence(t *testing.T) {
	source := urlReporterImageA + "\n" + urlReporterImageA

	_, reports := convertWithReports(t, source)

	expected := []URLReport{
		{Element: "img", Attribute: "src", URL: "/a.png"},
		{Element: "img", Attribute: "src", URL: "/a.png"},
	}
	assert.Equal(t, expected, reports)
}

func TestURLReporterReportsInlineImageInMultilineParagraphOnce(t *testing.T) {
	source := `first line
second ` + urlReporterImageP + ` line`

	_, reports := convertWithReports(t, source)

	expected := []URLReport{
		{Element: "img", Attribute: "src", URL: "/p.png"},
	}
	assert.Equal(t, expected, reports)
}

func TestURLReporterReportsImageInGlobal(t *testing.T) {
	_, reports := convertWithReports(t, `{{ G }}`,
		WithGlobalComponent("G", []byte(urlReporterImageA)),
	)

	expected := []URLReport{
		{Element: "img", Attribute: "src", URL: "/a.png"},
	}
	assert.Equal(t, expected, reports)
}

func TestURLReporterSetTwice(t *testing.T) {
	fn := func(URLReport) {}
	_, err := New().Convert([]byte("x"), io.Discard, WithURLReporter(fn), WithURLReporter(fn))
	require.Error(t, err)

	var compErr *ComponoError
	require.ErrorAs(t, err, &compErr)
	assert.Equal(t, ErrURLReporterAlreadySet, compErr.Code)
}

func TestURLReporterNil(t *testing.T) {
	source := urlReporterImageA + "\n" + urlReporterImageA

	var without bytes.Buffer
	_, err := New().Convert([]byte(source), &without)
	require.NoError(t, err)

	var withNil bytes.Buffer
	_, err = New().Convert([]byte(source), &withNil, WithURLReporter(nil))
	require.NoError(t, err)

	assert.Equal(t, without.String(), withNil.String())
}

func TestURLReporterNotAllowedInGlobal(t *testing.T) {
	fn := func(URLReport) {}
	_, err := New().Convert([]byte("{{ G }}"), io.Discard,
		WithGlobalComponent("G", []byte("x"), WithURLReporter(fn)),
	)

	require.Error(t, err)
	var compErr *ComponoError
	require.ErrorAs(t, err, &compErr)
	assert.Equal(t, ErrConversionOptionInGlobal, compErr.Code)
}