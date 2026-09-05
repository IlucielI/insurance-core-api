package config

import "os"

type Config struct {
	AppName  string
	AppEnv   string
	HTTPPort string
	Version  string
	GitHash  string
}

func Load() Config {
	return Config{
		AppName:  getEnv("APP_NAME", "insurance-core-api"),
		AppEnv:   getEnv("APP_ENV", "development"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		Version:  getEnv("APP_VERSION", Version),
		GitHash:  getEnv("GIT_HASH", GitHash),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
