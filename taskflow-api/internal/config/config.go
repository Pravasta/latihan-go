package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()

	if err != nil {
		log.Println("[Main] No .env file found, relying on environment variables")
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

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" envDefault:"my-secret"`
	Expire string `env:"JWT_EXPIRE" envDefault:"24h"`
}
