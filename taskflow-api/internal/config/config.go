package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Env      EnvConfig
	Logger   LoggerConfig
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()

	if err != nil {
		slog.Warn("[Config] No .env file found, relying on environment variables", "error", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "taskflow"),
			Password: getEnv("DB_PASSWORD", "taskflow"),
			Name:     getEnv("DB_NAME", "taskflow"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "my-secret"),
			Expire: getEnv("JWT_EXPIRE", "24h"),
		},
		Env: EnvConfig{
			Environment: getEnv("APP_ENV", "development"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type ServerConfig struct {
	Port string `env:"APP_PORT" envDefault:"8080"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER" envDefault:"taskflow"`
	Password string `env:"DB_PASSWORD" envDefault:"taskflow"`
	Name     string `env:"DB_NAME" envDefault:"taskflow"`
}

type EnvConfig struct {
	Environment string `env:"APP_ENV" envDefault:"development"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" envDefault:"my-secret"`
	Expire string `env:"JWT_EXPIRE" envDefault:"24h"`
}

type LoggerConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"debug"`
	Format string `env:"LOG_FORMAT" envDefault:"text"`
}
