package config

import (
	os "os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	createTempConfig := func(t *testing.T, content string) string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yml")
		err := os.WriteFile(path, []byte(content), 0644)
		require.NoError(t, err)
		return path
	}

	testCases := []struct {
		name        string
		setup       func(t *testing.T)
		checkResult func(t *testing.T, cfg *Config, err error)
	}{
		{
			name: "Success: Valid config and all envs set",
			setup: func(t *testing.T) {
				yamlContent := "env: test\nlog_path: /tmp/log.log\nhttp_server:\n  address: \"localhost:8080\"\nauth:\n  token_ttl: 30m"
				configPath := createTempConfig(t, yamlContent)
				t.Setenv("CONFIG_PATH", configPath)
				t.Setenv("DB_PASSWORD", "test_pass")
				t.Setenv("JWT_SECRET", "test_secret")
				t.Setenv("GOOGLE_CLIENT_ID", "test_google_id")
				t.Setenv("ENV", "prod") // This should override yaml
			},
			checkResult: func(t *testing.T, cfg *Config, err error) {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				require.Equal(t, "prod", cfg.Env)
				require.Equal(t, "/tmp/log.log", cfg.LogPath)
				require.Equal(t, "localhost:8080", cfg.Address)
				require.Equal(t, "5432", cfg.DBPort)
				require.Equal(t, "test_pass", cfg.DBPassword)
				require.Equal(t, 30*time.Minute, cfg.TokenTTL)
			},
		},
		{
			name: "Failure: CONFIG_PATH not set",
			setup: func(t *testing.T) {
				// Unset is not available, so we set it to empty
				t.Setenv("CONFIG_PATH", "")
			},
			checkResult: func(t *testing.T, cfg *Config, err error) {
				require.Error(t, err)
				require.Nil(t, cfg)
				require.Contains(t, err.Error(), "CONFIG_PATH environment variable is not set")
			},
		},
		{
			name: "Failure: Config file does not exist",
			setup: func(t *testing.T) {
				t.Setenv("CONFIG_PATH", "/path/to/non/existent/file.yml")
			},
			checkResult: func(t *testing.T, cfg *Config, err error) {
				require.Error(t, err)
				require.Nil(t, cfg)
				require.Contains(t, err.Error(), "config file does not exist at:")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Keep track of original env vars to restore them
			originalEnv := os.Environ()
			t.Cleanup(func() {
				os.Clearenv()
				for _, env := range originalEnv {
					parts := strings.SplitN(env, "=", 2)
					err := os.Setenv(parts[0], parts[1])
					if err != nil {
						return
					}
				}
			})

			tc.setup(t)
			cfg, err := Load()
			tc.checkResult(t, cfg, err)
		})
	}
}
