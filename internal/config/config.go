package config

import (
	"errors"
	"os"
)

type Config struct {
	AppName     string
	AppEnv      string
	HTTPPort    string
	Version     string
	GitHash     string
	DatabaseURL string
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return Config{
		AppName:     getEnv("APP_NAME", "insurance-core-api"),
		AppEnv:      getEnv("APP_ENV", "development"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		Version:     getEnv("APP_VERSION", Version),
		GitHash:     getEnv("GIT_HASH", GitHash),
		DatabaseURL: databaseURL,
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
