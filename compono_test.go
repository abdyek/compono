package compono

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/umono-cms/compono/logger"
)

type componoTestSuite struct {
	suite.Suite
}

type invalidContextKeyNotation struct {
	Title string `compono:"invalid_key"`
}

func (s *componoTestSuite) TestGolden() {
	inputFiles, err := filepath.Glob("testdata/input/*.comp")
	require.Nil(s.T(), err)
	require.NotEmpty(s.T(), inputFiles, "no .comp files found")

	for _, inputPath := range inputFiles {
		name := filepath.Base(inputPath)
		input, err := os.ReadFile(inputPath)
		require.Nil(s.T(), err)

		globalFiles, err := filepath.Glob("testdata/input/global/" + strings.TrimSuffix(name, ".comp") + "/*.comp")
		require.Nil(s.T(), err)

		comp := New()
		comp.Logger().SetLogLevel(logger.All)

		opts := []ConvertOption{}
		for _, gPath := range globalFiles {
			globalCompName := filepath.Base(gPath)
			globalInput, err := os.ReadFile(gPath)
			require.Nil(s.T(), err)

			opts = append(opts, WithGlobalComponent(strings.TrimSuffix(globalCompName, ".comp"), []byte(strings.TrimSpace(string(globalInput)))))
		}

		contextPath := filepath.Join("testdata/input/context", strings.TrimSuffix(name, ".comp")+".json")
		if _, err := os.Stat(contextPath); err == nil {
			contextValues, err := readContextFixture(contextPath)
			require.Nil(s.T(), err)
			opts = append(opts, WithContext(contextValues))
		}

		var buf bytes.Buffer
		err = comp.Convert([]byte(strings.TrimSpace(string(input))), &buf, opts...)
		assert.Nil(s.T(), err)

		goldenPath := filepath.Join(
			"testdata/output",
			strings.TrimSuffix(name, ".comp")+".golden",
		)

		golden, err := os.ReadFile(goldenPath)
		require.Nil(s.T(), err, "golden file missing")

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "from %s", inputPath)
	}
}

func (s *componoTestSuite) TestGoldenForWithGlobalComponent() {
	inputFiles, err := filepath.Glob("testdata/global_input/*.comp")
	require.Nil(s.T(), err)
	require.NotEmpty(s.T(), inputFiles, "no .comp files found")

	for _, inputPath := range inputFiles {
		name := filepath.Base(inputPath)
		input, err := os.ReadFile(inputPath)
		require.Nil(s.T(), err)

		globalFiles, err := filepath.Glob("testdata/global_input/global/" + strings.TrimSuffix(name, ".comp") + "/*.comp")
		require.Nil(s.T(), err)

		comp := New()
		comp.Logger().SetLogLevel(logger.All)

		opts := []ConvertOption{}
		for _, gPath := range globalFiles {
			globalCompName := filepath.Base(gPath)
			globalInput, err := os.ReadFile(gPath)
			require.Nil(s.T(), err)

			opts = append(opts, WithGlobalComponent(strings.TrimSuffix(globalCompName, ".comp"), []byte(strings.TrimSpace(string(globalInput)))))
		}
		opts = append(opts, WithGlobalComponent(strings.TrimSuffix(name, ".comp"), []byte(strings.TrimSpace(string(input)))))

		var buf bytes.Buffer
		err = comp.Convert(
			[]byte(`{{ `+strings.TrimSuffix(name, ".comp")+` }}`),
			&buf,
			opts...,
		)
		assert.Nil(s.T(), err)

		goldenPath := filepath.Join(
			"testdata/global_output",
			strings.TrimSuffix(name, ".comp")+".golden",
		)

		golden, err := os.ReadFile(goldenPath)
		require.Nil(s.T(), err, "golden file missing")

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "from %s", inputPath)
	}
}

func (s *componoTestSuite) TestConvertWithContextErrUnsupportedType() {
	compono := New()

	var buf bytes.Buffer
	err := compono.Convert([]byte("context"), &buf, WithContext(map[string]any{
		"value": func() {},
	}))

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrUnsupportedType, compErr.Code)
	assert.Contains(s.T(), compErr.Message, "unsupported context value type")
}

func (s *componoTestSuite) TestConvertWithContextErrUnsupportedKeyNotation() {
	compono := New()

	var buf bytes.Buffer
	err := compono.Convert([]byte("context"), &buf, WithContext(map[string]any{
		"value": invalidContextKeyNotation{Title: "Hello"},
	}))

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrUnsupportedKeyNotation, compErr.Code)
	assert.Contains(s.T(), compErr.Message, `invalid compono struct tag "invalid_key"`)
}

const (
	errorBlockHead  = `<compono-error-block><template shadowrootmode="closed">`
	errorInlineHead = `<compono-error-inline><template shadowrootmode="closed">`
	errorBlockTail  = `</template></compono-error-block>`
	errorInlineTail = `</template></compono-error-inline>`

	unknownFooBlockBody  = `<div class="title">Unknown component</div><div class="description">The component <strong>FOO</strong> is not defined or not registered.</div>`
	unknownFooInlineBody = `<span class="title">Unknown component</span><span class="description">The component <strong>FOO</strong> is not defined or not registered.</span>`
	unknownMissingBlock  = `<div class="title">Unknown component</div><div class="description">The component <strong>MISSING</strong> is not defined or not registered.</div>`
)

func (s *componoTestSuite) TestWithErrorStylesheetBlockError() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf, WithErrorStylesheet("/_umono/error.css"))
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		errorBlockHead+`<link rel="stylesheet" href="/_umono/error.css">`+unknownFooBlockBody+errorBlockTail,
		buf.String(),
	)
}

func (s *componoTestSuite) TestWithErrorStylesheetInlineErrorInParagraph() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`Hello {{ FOO }} world`), &buf, WithErrorStylesheet("/_umono/error.css"))
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		`<p>Hello `+errorInlineHead+`<link rel="stylesheet" href="/_umono/error.css">`+unknownFooInlineBody+errorInlineTail+` world</p>`,
		buf.String(),
	)
}

func (s *componoTestSuite) TestWithoutErrorStylesheetHasNoLink() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf)
	require.Nil(s.T(), err)

	assert.Equal(s.T(), errorBlockHead+unknownFooBlockBody+errorBlockTail, buf.String())
	assert.NotContains(s.T(), buf.String(), "<link")
}

func (s *componoTestSuite) TestWithErrorStylesheetEmptyStringHasNoLink() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf, WithErrorStylesheet(""))
	require.Nil(s.T(), err)

	assert.Equal(s.T(), errorBlockHead+unknownFooBlockBody+errorBlockTail, buf.String())
	assert.NotContains(s.T(), buf.String(), "<link")

	buf.Reset()
	err = New().Convert([]byte(`Hello {{ FOO }}`), &buf, WithErrorStylesheet(""))
	require.Nil(s.T(), err)

	assert.Equal(s.T(), `<p>Hello `+errorInlineHead+unknownFooInlineBody+errorInlineTail+`</p>`, buf.String())
	assert.NotContains(s.T(), buf.String(), "<link")
}

func (s *componoTestSuite) TestWithErrorStylesheetEscapesHref() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf, WithErrorStylesheet(`/a"b&c`))
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		errorBlockHead+`<link rel="stylesheet" href="/a&#34;b&amp;c">`+unknownFooBlockBody+errorBlockTail,
		buf.String(),
	)
}

func (s *componoTestSuite) TestWithErrorStylesheetDoesNotValidateURL() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf, WithErrorStylesheet(`not a url <x>'`))
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		errorBlockHead+`<link rel="stylesheet" href="not a url &lt;x&gt;&#39;">`+unknownFooBlockBody+errorBlockTail,
		buf.String(),
	)
}

func (s *componoTestSuite) TestWithErrorStylesheetAppliesToEveryError() {
	var buf bytes.Buffer
	err := New().Convert([]byte("{{ FOO }}\n\nHello {{ FOO }}"), &buf, WithErrorStylesheet("/e.css"))
	require.Nil(s.T(), err)

	link := `<link rel="stylesheet" href="/e.css">`
	assert.Equal(s.T(),
		errorBlockHead+link+unknownFooBlockBody+errorBlockTail+
			`<p>Hello `+errorInlineHead+link+unknownFooInlineBody+errorInlineTail+`</p>`,
		buf.String(),
	)
}

func (s *componoTestSuite) TestConvertWithErrorStylesheetErrErrorStylesheetAlreadySet() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf,
		WithErrorStylesheet("/a.css"),
		WithErrorStylesheet("/b.css"),
	)

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrErrorStylesheetAlreadySet, compErr.Code)
	assert.Contains(s.T(), compErr.Message, "error stylesheet")
	assert.Empty(s.T(), buf.String())
}

func (s *componoTestSuite) TestConvertWithErrorStylesheetTwiceErrEvenWhenEmpty() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ FOO }}`), &buf,
		WithErrorStylesheet("/a.css"),
		WithErrorStylesheet(""),
	)

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrErrorStylesheetAlreadySet, compErr.Code)
}

func (s *componoTestSuite) TestWithErrorStylesheetIsPerConversion() {
	comp := New()

	var first bytes.Buffer
	err := comp.Convert([]byte(`{{ FOO }}`), &first, WithErrorStylesheet("/a.css"))
	require.Nil(s.T(), err)
	assert.Contains(s.T(), first.String(), `href="/a.css"`)

	var second bytes.Buffer
	err = comp.Convert([]byte(`{{ FOO }}`), &second)
	require.Nil(s.T(), err)
	assert.Equal(s.T(), errorBlockHead+unknownFooBlockBody+errorBlockTail, second.String())

	var third bytes.Buffer
	err = comp.Convert([]byte(`{{ FOO }}`), &third, WithErrorStylesheet("/b.css"))
	require.Nil(s.T(), err)
	assert.Contains(s.T(), third.String(), `href="/b.css"`)
	assert.NotContains(s.T(), third.String(), `/a.css`)
}

func (s *componoTestSuite) TestWithErrorStylesheetAppliesToWithGlobalComponentError() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ GREETING }}`), &buf,
		WithGlobalComponent("GREETING", []byte(`{{ MISSING }}`)),
		WithErrorStylesheet("/_umono/error.css"),
	)
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		errorBlockHead+`<link rel="stylesheet" href="/_umono/error.css">`+unknownMissingBlock+errorBlockTail,
		buf.String(),
	)
}

func (s *componoTestSuite) TestWithErrorStylesheetAppliesToInlineErrorInGlobalComponent() {
	var buf bytes.Buffer
	err := New().Convert([]byte(`{{ GREETING }}`), &buf,
		WithErrorStylesheet("/_umono/error.css"),
		WithGlobalComponent("GREETING", []byte(`Hello {{ MISSING }}`)),
	)
	require.Nil(s.T(), err)

	assert.Equal(s.T(),
		`<p>Hello `+errorInlineHead+`<link rel="stylesheet" href="/_umono/error.css">`+
			`<span class="title">Unknown component</span><span class="description">The component <strong>MISSING</strong> is not defined or not registered.</span>`+
			errorInlineTail+`</p>`,
		buf.String(),
	)
}

func (s *componoTestSuite) TestGoldenForScopes() {
	caseDirs, err := filepath.Glob("testdata/scope/*")
	require.Nil(s.T(), err)
	require.NotEmpty(s.T(), caseDirs, "no scope test cases found")

	for _, caseDir := range caseDirs {
		caseName := filepath.Base(caseDir)

		input, err := os.ReadFile(filepath.Join(caseDir, "input.comp"))
		require.Nil(s.T(), err)

		golden, err := os.ReadFile(filepath.Join(caseDir, "output.golden"))
		require.Nil(s.T(), err)

		comp := New()
		comp.Logger().SetLogLevel(logger.All)

		opts := buildScopeOpts(s.T(), filepath.Join(caseDir, "global"))

		var buf bytes.Buffer
		err = comp.Convert([]byte(strings.TrimSpace(string(input))), &buf, opts...)
		assert.Nil(s.T(), err)

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "case %s", caseName)
	}
}

func buildScopeOpts(t *testing.T, dir string) []ConvertOption {
	t.Helper()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	var opts []ConvertOption
	if _, err := os.Stat(filepath.Join(dir, ".isolated")); err == nil {
		opts = append(opts, WithIsolatedScope())
	}

	compFiles, err := filepath.Glob(filepath.Join(dir, "*.comp"))
	require.Nil(t, err)

	for _, cPath := range compFiles {
		name := strings.TrimSuffix(filepath.Base(cPath), ".comp")
		src, err := os.ReadFile(cPath)
		require.Nil(t, err)

		subOpts := buildScopeOpts(t, filepath.Join(dir, name))

		opts = append(opts, WithGlobalComponent(name, []byte(strings.TrimSpace(string(src))), subOpts...))
	}
	return opts
}

func (s *componoTestSuite) TestErrConversionOptionInGlobal() {
	err := New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithContext(map[string]any{"k": "v"})),
	)
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithErrorStylesheet("/a.css")),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithContext(nil)),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithErrorStylesheet("")),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`),
			WithGlobalComponent("SUB", []byte(`sub`), WithContext(map[string]any{})),
		),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithAttributeHook(nil)),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)
}

func (s *componoTestSuite) TestErrIsolatedScopeInConvert() {
	err := New().Convert([]byte(`hello`), io.Discard, WithIsolatedScope())
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrIsolatedScopeInConvert, compErr.Code)
}

func (s *componoTestSuite) TestErrDuplicateSubComponent() {
	err := New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`{{ A }}`),
			WithGlobalComponent("A", []byte(`first`)),
			WithGlobalComponent("A", []byte(`second`)),
		),
	)
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateSubComponent, compErr.Code)

	err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`{{ H }}`),
			WithGlobalComponent("H", []byte(`{{ A }}`),
				WithGlobalComponent("A", []byte(`first`)),
				WithGlobalComponent("A", []byte(`second`)),
			),
		),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateSubComponent, compErr.Code)
}

func (s *componoTestSuite) TestErrDuplicateGlobalComponent() {
	err := New().Convert([]byte(`{{ A }}`), io.Discard,
		WithGlobalComponent("A", []byte(`first`)),
		WithGlobalComponent("A", []byte(`second`)),
	)
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateGlobalComponent, compErr.Code)

	err = New().Convert([]byte(`{{ A }}`), io.Discard,
		WithGlobalComponent("A", []byte(`first`)),
		WithGlobalComponent("B", []byte(`other`)),
		WithGlobalComponent("A", []byte(`second`)),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateGlobalComponent, compErr.Code)

	var buf bytes.Buffer
	err = New().Convert([]byte("{{ A }}\n\n{{ G }}"), &buf,
		WithGlobalComponent("A", []byte(`root`)),
		WithGlobalComponent("G", []byte(`{{ A }}`),
			WithGlobalComponent("A", []byte(`inner`)),
		),
	)
	require.Nil(s.T(), err)
	assert.Equal(s.T(), `<p>root</p><p>inner</p>`, buf.String())
}

func (s *componoTestSuite) TestIsolatedScopeTwiceSameAsOnce() {
	var once, twice bytes.Buffer

	c1 := New()
	err := c1.Convert([]byte(`{{ G }}`), &once,
		WithGlobalComponent("R", []byte(`root`)),
		WithGlobalComponent("G", []byte(`{{ R }}`), WithIsolatedScope()),
	)
	require.Nil(s.T(), err)

	c2 := New()
	err = c2.Convert([]byte(`{{ G }}`), &twice,
		WithGlobalComponent("R", []byte(`root`)),
		WithGlobalComponent("G", []byte(`{{ R }}`), WithIsolatedScope(), WithIsolatedScope()),
	)
	require.Nil(s.T(), err)

	assert.Equal(s.T(), once.String(), twice.String())
	assert.Contains(s.T(), once.String(), "Unknown component")
}

func TestComponoTestSuite(t *testing.T) {
	suite.Run(t, new(componoTestSuite))
}

func readContextFixture(path string) (map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	decoder.UseNumber()

	values := map[string]any{}
	if err := decoder.Decode(&values); err != nil {
		return nil, err
	}

	return normalizeJSONValue(values).(map[string]any), nil
}

func normalizeJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeJSONValue(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, normalizeJSONValue(item))
		}
		return result
	case json.Number:
		if i, err := strconv.ParseInt(string(typed), 10, 64); err == nil {
			return i
		}
		return typed
	default:
		return value
	}
}
