package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"local"`
	LogPath    string `yaml:"log_path" env:"LOG_PATH" env-default:"./logs/skrepka.log"`
	HTTPServer `yaml:"http_server"`
	DB         `yaml:"postgres"`
	Auth       `yaml:"auth"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:4000"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type DB struct {
	DBHost     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	DBPort     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	DBUser     string `yaml:"user" env:"DB_USER" env-default:"skrepka_user"`
	DBPassword string `yaml:"password" env:"DB_PASSWORD" env-required:"true"`
	DBName     string `yaml:"db_name" env:"DB_NAME" env-default:"skrepka_db"`
	SSLMode    string `yaml:"sslmode" env-default:"disable"`
}

type Auth struct {
	JWTSecret      string        `yaml:"jwt_secret" env:"JWT_SECRET" env-required:"true"`
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
	GoogleClientID string        `yaml:"google_client_id" env:"GOOGLE_CLIENT_ID" env-required:"true"`
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		configPath := os.Getenv("CONFIG_PATH")
		if configPath == "" {
			log.Fatal("CONFIG_PATH environment variable is not set")
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			log.Fatalf("config file does not exist at: %s", configPath)
		}

		var cfg Config

		if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
			log.Fatalf("cannot read config: %s", err)
		}

		instance = &cfg
	})

	return instance
}

// ResetInstanceForTesting resets global config instance.
// SHOULD BE USED ONLY IN TESTING
func ResetInstanceForTesting() {
	instance = nil
	once = sync.Once{}
}
