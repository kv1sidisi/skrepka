package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupLogger(t *testing.T) {
	testCases := []struct {
		name          string
		env           string
		expectedLevel slog.Level
		mustContain   string
	}{
		{
			name:          "local env",
			env:           "local",
			expectedLevel: slog.LevelDebug,
			mustContain:   `level=DEBUG msg="test debug message"`,
		},
		{
			name:          "dev env",
			env:           "dev",
			expectedLevel: slog.LevelDebug,
			mustContain:   `"level":"DEBUG","msg":"test debug message"`,
		},
		{
			name:          "prod env",
			env:           "prod",
			expectedLevel: slog.LevelInfo,
			mustContain:   `"level":"INFO","msg":"test info message"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			logger := SetupLogger(tc.env, &buf)

			require.True(t, logger.Enabled(nil, tc.expectedLevel))
			require.False(t, logger.Enabled(nil, tc.expectedLevel-1))

			if tc.expectedLevel == slog.LevelDebug {
				logger.Debug("test debug message")
			} else {
				logger.Info("test info message")
			}

			output := buf.String()
			require.True(t, strings.Contains(output, tc.mustContain), "log output '%s' does not contain expected string '%s'", output, tc.mustContain)
		})
	}
}
