package loggingx_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.infratographer.com/x/loggingx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

var expectedMsg = "ohno"

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name       string
		cfg        loggingx.Config
		expectedST bool
	}{
		{
			name:       "With all disabled",
			cfg:        loggingx.Config{Debug: false, Pretty: false, DisableStacktrace: true},
			expectedST: false,
		},
		{
			name:       "With stacktrace disabled",
			cfg:        loggingx.Config{Debug: false, Pretty: false, DisableStacktrace: true},
			expectedST: false,
		},
		{
			name:       "With all enabled",
			cfg:        loggingx.Config{Debug: true, Pretty: true, DisableStacktrace: false},
			expectedST: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := loggingx.InitLogger("test", tt.cfg)

			// Replace the logger's core with the observer core
			core, recorded := observer.New(zapcore.DebugLevel)
			logger = logger.Desugar().WithOptions(zap.WrapCore(func(_ zapcore.Core) zapcore.Core {
				return core
			})).Sugar()

			// Log an error message
			logger.Error(expectedMsg)

			// Assert that the message was logged
			require.Equal(t, 1, recorded.Len(), "Expected one log entry")
			allLogs := recorded.All()
			require.Equal(t, expectedMsg, allLogs[0].Message, "Logged message does not match expected")

			// Assert that the stacktrace was logged as expected
			if tt.expectedST {
				require.NotEmpty(t, allLogs[0].Stack, "Expected stacktrace to be logged")
			} else {
				require.Empty(t, allLogs[0].Stack, "Expected no stacktrace to be logged")
			}
		})
	}
}
