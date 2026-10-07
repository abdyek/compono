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
		diags, err := comp.Convert([]byte(strings.TrimSpace(string(input))), &buf, opts...)
		assert.Nil(s.T(), err)

		baseName := strings.TrimSuffix(name, ".comp")
		goldenPath := filepath.Join(
			"testdata/output",
			baseName+".golden",
		)

		golden, err := os.ReadFile(goldenPath)
		require.Nil(s.T(), err, "golden file missing")

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "from %s", inputPath)
		assert.Equal(s.T(), readDiagnosticsGolden(s.T(), filepath.Join("testdata/output", baseName+".diag")), formatDiagnostics(diags), "from %s", inputPath)
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
		diags, err := comp.Convert(
			[]byte(`{{ `+strings.TrimSuffix(name, ".comp")+` }}`),
			&buf,
			opts...,
		)
		assert.Nil(s.T(), err)

		baseName := strings.TrimSuffix(name, ".comp")
		goldenPath := filepath.Join(
			"testdata/global_output",
			baseName+".golden",
		)

		golden, err := os.ReadFile(goldenPath)
		require.Nil(s.T(), err, "golden file missing")

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "from %s", inputPath)
		assert.Equal(s.T(), readDiagnosticsGolden(s.T(), filepath.Join("testdata/global_output", baseName+".diag")), formatDiagnostics(diags), "from %s", inputPath)
	}
}

func (s *componoTestSuite) TestConvertWithContextErrUnsupportedType() {
	compono := New()

	var buf bytes.Buffer
	_, err := compono.Convert([]byte("context"), &buf, WithContext(map[string]any{
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
	_, err := compono.Convert([]byte("context"), &buf, WithContext(map[string]any{
		"value": invalidContextKeyNotation{Title: "Hello"},
	}))

	require.Error(s.T(), err)

	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrUnsupportedKeyNotation, compErr.Code)
	assert.Contains(s.T(), compErr.Message, `invalid compono struct tag "invalid_key"`)
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
		diags, err := comp.Convert([]byte(strings.TrimSpace(string(input))), &buf, opts...)
		assert.Nil(s.T(), err)

		assert.Equal(s.T(), strings.TrimSpace(string(golden)), buf.String(), "case %s", caseName)
		assert.Equal(s.T(), readDiagnosticsGolden(s.T(), filepath.Join(caseDir, "output.diag")), formatDiagnostics(diags), "case %s", caseName)
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
	_, err := New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithContext(map[string]any{"k": "v"})),
	)
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	_, err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithContext(nil)),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	_, err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`),
			WithGlobalComponent("SUB", []byte(`sub`), WithContext(map[string]any{})),
		),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)

	_, err = New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`hi`), WithAttributeHook(nil)),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrConversionOptionInGlobal, compErr.Code)
}

func (s *componoTestSuite) TestErrIsolatedScopeInConvert() {
	_, err := New().Convert([]byte(`hello`), io.Discard, WithIsolatedScope())
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrIsolatedScopeInConvert, compErr.Code)
}

func (s *componoTestSuite) TestErrDuplicateSubComponent() {
	_, err := New().Convert([]byte(`{{ G }}`), io.Discard,
		WithGlobalComponent("G", []byte(`{{ A }}`),
			WithGlobalComponent("A", []byte(`first`)),
			WithGlobalComponent("A", []byte(`second`)),
		),
	)
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateSubComponent, compErr.Code)

	_, err = New().Convert([]byte(`{{ G }}`), io.Discard,
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
	_, err := New().Convert([]byte(`{{ A }}`), io.Discard,
		WithGlobalComponent("A", []byte(`first`)),
		WithGlobalComponent("A", []byte(`second`)),
	)
	require.Error(s.T(), err)
	var compErr *ComponoError
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateGlobalComponent, compErr.Code)

	_, err = New().Convert([]byte(`{{ A }}`), io.Discard,
		WithGlobalComponent("A", []byte(`first`)),
		WithGlobalComponent("B", []byte(`other`)),
		WithGlobalComponent("A", []byte(`second`)),
	)
	require.Error(s.T(), err)
	require.ErrorAs(s.T(), err, &compErr)
	assert.Equal(s.T(), ErrDuplicateGlobalComponent, compErr.Code)

	var buf bytes.Buffer
	_, err = New().Convert([]byte("{{ A }}\n\n{{ G }}"), &buf,
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
	onceDiags, err := c1.Convert([]byte(`{{ G }}`), &once,
		WithGlobalComponent("R", []byte(`root`)),
		WithGlobalComponent("G", []byte(`{{ R }}`), WithIsolatedScope()),
	)
	require.Nil(s.T(), err)

	c2 := New()
	twiceDiags, err := c2.Convert([]byte(`{{ G }}`), &twice,
		WithGlobalComponent("R", []byte(`root`)),
		WithGlobalComponent("G", []byte(`{{ R }}`), WithIsolatedScope(), WithIsolatedScope()),
	)
	require.Nil(s.T(), err)

	assert.Equal(s.T(), once.String(), twice.String())
	assert.Equal(s.T(), onceDiags, twiceDiags)
	require.Len(s.T(), onceDiags, 1)
	assert.Equal(s.T(), CodeUnknownComponent, onceDiags[0].Code)
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
