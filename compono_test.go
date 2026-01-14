package compono

import (
	"bytes"
	"os"
	"path/filepath"
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

		var buf bytes.Buffer
		err = comp.Convert([]byte(strings.TrimSpace(string(input))), &buf)
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

func (s *componoTestSuite) TestGoldenForConvertGlobalComponent() {
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
		err = comp.ConvertGlobalComponent(strings.TrimSuffix(name, ".comp"), []byte(strings.TrimSpace(string(input))), &buf)
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

func TestComponoTestSuite(t *testing.T) {
	suite.Run(t, new(componoTestSuite))
}
