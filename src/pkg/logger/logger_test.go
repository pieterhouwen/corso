package logger_test

import (
	"testing"

	"github.com/alcionai/clues"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/alcionai/corso/src/internal/tester"
	"github.com/alcionai/corso/src/pkg/logger"
)

type LoggerUnitSuite struct {
	tester.Suite
}

func TestLoggerUnitSuite(t *testing.T) {
	suite.Run(t, &LoggerUnitSuite{Suite: tester.NewUnitSuite(t)})
}

func (suite *LoggerUnitSuite) TestAddLoggingFlags() {
	t := suite.T()

	logger.DebugAPIFV = false
	logger.ReadableLogsFV = false

	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			assert.True(t, logger.DebugAPIFV, logger.DebugAPIFN)
			assert.True(t, logger.ReadableLogsFV, logger.ReadableLogsFN)
			assert.Equal(t, logger.LLError, logger.LogLevelFV, logger.LogLevelFN)
			assert.Equal(t, logger.PIIMask, logger.SensitiveInfoFV, logger.SensitiveInfoFN)
			// empty assertion here, instead of matching "log-file", because the LogFile
			// var isn't updated by running the command (this is expected and correct),
			// while the logFileFV remains unexported.
			assert.Empty(t, logger.LogFile, logger.LogFileFN)
		},
	}

	logger.AddLoggingFlags(cmd)

	// Test arg parsing for few args
	cmd.SetArgs([]string{
		"test",
		"--" + logger.DebugAPIFN,
		"--" + logger.LogFileFN, "log-file",
		"--" + logger.LogLevelFN, logger.LLError,
		"--" + logger.ReadableLogsFN,
		"--" + logger.SensitiveInfoFN, logger.PIIMask,
	})

	err := cmd.Execute()
	require.NoError(t, err, clues.ToCore(err))
}

func (suite *LoggerUnitSuite) TestPreloadLoggingFlags() {
	t := suite.T()

	logger.DebugAPIFV = false
	logger.ReadableLogsFV = false

	args := []string{
		"--" + logger.DebugAPIFN,
		"--" + logger.LogFileFN, "log-file",
		"--" + logger.LogLevelFN, logger.LLError,
		"--" + logger.ReadableLogsFN,
		"--" + logger.SensitiveInfoFN, logger.PIIMask,
	}

	settings := logger.PreloadLoggingFlags(args)

	assert.True(t, logger.DebugAPIFV, logger.DebugAPIFN)
	assert.True(t, logger.ReadableLogsFV, logger.ReadableLogsFN)
	assert.Equal(t, "log-file", settings.File, "settings.File")
	assert.Equal(t, logger.LLError, settings.Level, "settings.Level")
	assert.Equal(t, logger.PIIMask, settings.PIIHandling, "settings.PIIHandling")
}