package compono

import (
	"bytes"
	"encoding/json"
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

		for _, gPath := range globalFiles {
			globalCompName := filepath.Base(gPath)
			globalInput, err := os.ReadFile(gPath)
			require.Nil(s.T(), err)

			err = comp.RegisterGlobalComponent(strings.TrimSuffix(globalCompName, ".comp"), []byte(strings.TrimSpace(string(globalInput))))
			assert.Nil(s.T(), err)
		}

		opts := []ConvertOption{}
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

		for _, gPath := range globalFiles {
			globalCompName := filepath.Base(gPath)
			globalInput, err := os.ReadFile(gPath)
			require.Nil(s.T(), err)

			err = comp.RegisterGlobalComponent(strings.TrimSuffix(globalCompName, ".comp"), []byte(strings.TrimSpace(string(globalInput))))
			assert.Nil(s.T(), err)
		}

		var buf bytes.Buffer
		err = comp.Convert(
			[]byte(`{{ `+strings.TrimSuffix(name, ".comp")+` }}`),
			&buf,
			WithGlobalComponent(strings.TrimSuffix(name, ".comp"), []byte(strings.TrimSpace(string(input)))),
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

func (s *componoTestSuite) TestUnregisterGlobalComponent() {
	compono := New().(*compono)
	err := compono.RegisterGlobalComponent("SAY_HELLO", []byte("# Hello"))
	require.Nil(s.T(), err)
	err = compono.UnregisterGlobalComponent("SAY_HELLO")
	require.Nil(s.T(), err)
	assert.Equal(s.T(), 0, len(compono.globalWrapper.Children()))
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
