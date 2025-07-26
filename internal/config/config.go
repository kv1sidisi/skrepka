package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config stores all settings application needs to start and run correctly.
type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	LogPath    string `yaml:"log_path" env:"LOG_PATH" env-default:"./logs/skrepka.log"`
	HTTPServer `yaml:"http_server"`
	DB         `yaml:"postgres"`
	Auth       `yaml:"auth"`
}

// HTTPServer stores settings for web server, like which address and port to use.
type HTTPServer struct {
	Address     string        `yaml:"address" env-default:":4000"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// DB stores settings to connect to postgres database.
type DB struct {
	DBHost     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	DBPort     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	DBUser     string `yaml:"user" env:"DB_USER" env-default:"skrepka_user"`
	DBPassword string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	DBName     string `yaml:"db_name" env:"DB_NAME" env-default:"skrepka_db"`
	SSLMode    string `yaml:"sslmode" env-default:"disable"`
}

// Auth stores settings for authentication, like secrets for signing tokens.
type Auth struct {
	JWTSecret      string        `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
	GoogleClientID string        `yaml:"google_client_id" env:"GOOGLE_CLIENT_ID" env-required:"true"`
}

// Load reads configuration from file and environment variables.
// Returns pointer to Config and error on failure.
func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist at: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("cannot read config: %w", err)
	}

	return &cfg, nil
}
