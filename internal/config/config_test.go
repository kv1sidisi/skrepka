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

// TestMain serves as a wrapper for the test suite, enabling a special execution mode
// to test functions that terminate the program, such as those using log.Fatal().
// Standard tests cannot catch os.Exit() calls, so this approach is necessary.
//
// The test is run in a separate child process by setting a specific environment variable.
// The main test process then analyzes the exit code and output of the child process.
func TestMain(m *testing.M) {
	if os.Getenv("GO_TEST_PROCESS") == "1" {
		// When in the child process, execute the function under test and exit.
		_ = MustLoad()
		return
	}
	// In the main process, run all other tests normally.
	m.Run()
}

// TestMustLoad covers all execution paths for the MustLoad function.
func TestMustLoad(t *testing.T) {
	// createTempConfig is a helper function to generate a temporary configuration file.
	// It uses t.TempDir() to ensure automatic cleanup after the test completes.
	createTempConfig := func(t *testing.T, content string) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yml")
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
		return path
	}

	// testCases defines a table of test scenarios for the MustLoad function.
	testCases := []struct {
		name             string
		setup            func(t *testing.T) []string
		checkResult      func(t *testing.T, cfg *Config)
		expectFatal      bool
		fatalMsgContains string
	}{
		// Test case 1: Verifies the successful loading of a valid configuration.
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
					// This environment variable should override the value from the YAML file.
					"ENV=prod",
				}
			},
			// checkResult validates the fields of the resulting Config struct.
			checkResult: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg)
				require.Equal(t, "prod", cfg.Env)
				require.Equal(t, "/tmp/log.log", cfg.LogPath)
				require.Equal(t, "localhost:8080", cfg.Address)
				require.Equal(t, "5432", cfg.Port)          // Checks default value.
				require.Equal(t, "test_pass", cfg.Password) // Checks required env var.
				require.Equal(t, 30*time.Minute, cfg.TokenTTL)
			},
		},
		// Test case 2: Verifies that the program terminates if CONFIG_PATH is not set.
		{
			name:             "Failure: CONFIG_PATH not set",
			setup:            func(t *testing.T) []string { return nil },
			expectFatal:      true,
			fatalMsgContains: "CONFIG_PATH environment variable is not set",
		},
		// Test case 3: Verifies that the program terminates if the config file does not exist.
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
			// For non-fatal test cases, run the function in the current process.
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

			// For fatal test cases, execute the test in a separate process.
			cmd := exec.Command(os.Args[0], "-test.run="+t.Name())
			cmd.Env = append(os.Environ(), "GO_TEST_PROCESS=1")
			cmd.Env = append(cmd.Env, tc.setup(t)...)

			output, err := cmd.CombinedOutput()

			// Assert that the command failed as expected.
			require.Error(t, err, "Expected command to fail, but it succeeded")
			var exitErr *exec.ExitError
			ok := errors.As(err, &exitErr)
			require.True(t, ok, "Expected error to be of type *exec.ExitError")
			require.False(t, exitErr.Success(), "Expected command to exit with a non-zero status")

			// Assert that the output contains the expected fatal error message.
			require.True(t, strings.Contains(string(output), tc.fatalMsgContains),
				"expected log output to contain '%s', but got '%s'", tc.fatalMsgContains, string(output),
			)
		})
	}
}
