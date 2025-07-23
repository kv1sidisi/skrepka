package config

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		_ = MustLoad()
		return
	}
	m.Run()
}

func TestMustLoad(t *testing.T) {
	createTempConfig := func(t *testing.T, content string) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yml")
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
		return path
	}

	testCases := []struct {
		name             string
		setup            func(t *testing.T) []string // Возвращает переменные окружения для дочернего процесса
		checkResult      func(t *testing.T, cfg *Config)
		expectFatal      bool
		fatalMsgContains string
	}{
		{
			name: "Success: Valid config and all envs set",
			setup: func(t *testing.T) []string {
				yamlContent := "env: test\nlog_path: /tmp/log.log\nhttp_server:\n  address: \"localhost:8080\"\nauth:\n  token_ttl: 30m"
				configPath := createTempConfig(t, yamlContent)
				return []string{
					"CONFIG_PATH=" + configPath,
					"DB_PASSWORD=test_pass",
					"JWT_SECRET=test_secret",
					"GOOGLE_CLIENT_ID=test_google_id",
					"ENV=prod",
				}
			},
			checkResult: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg)
				require.Equal(t, "prod", cfg.Env)
				require.Equal(t, "/tmp/log.log", cfg.LogPath)
				require.Equal(t, "localhost:8080", cfg.HTTPServer.Address)
				require.Equal(t, "5432", cfg.DB.Port)
				require.Equal(t, "test_pass", cfg.DB.Password)
				require.Equal(t, 30*time.Minute, cfg.Auth.TokenTTL)
			},
		},
		{
			name:             "Failure: CONFIG_PATH not set",
			setup:            func(t *testing.T) []string { return nil },
			expectFatal:      true,
			fatalMsgContains: "CONFIG_PATH environment variable is not set",
		},
		{
			name: "Failure: Config file does not exist",
			setup: func(t *testing.T) []string {
				return []string{"CONFIG_PATH=/path/to/non/existent/file.yml"}
			},
			expectFatal:      true,
			fatalMsgContains: "config file does not exist at:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.expectFatal {
				envs := tc.setup(t)
				for _, env := range envs {
					parts := strings.SplitN(env, "=", 2)
					t.Setenv(parts[0], parts[1])
				}
				cfg := MustLoad()
				if tc.checkResult != nil {
					tc.checkResult(t, cfg)
				}
				return
			}

			cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
			cmd.Env = append(os.Environ(), "GO_TEST_PROCESS=1")
			cmd.Env = append(cmd.Env, tc.setup(t)...)

			output, err := cmd.CombinedOutput()

			require.Error(t, err, "Expected command to fail, but it succeeded")
			var exitErr *exec.ExitError
			ok := errors.As(err, &exitErr)
			require.True(t, ok, "Expected error to be of type *exec.ExitError")
			require.False(t, exitErr.Success(), "Expected command to exit with a non-zero status")

			require.True(t, strings.Contains(string(output), tc.fatalMsgContains),
				"Expected log output to contain '%s', but got '%s'", tc.fatalMsgContains, string(output),
			)
		})
	}
}
