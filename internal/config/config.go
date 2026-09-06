package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName            string
	AppEnv             string
	HTTPPort           string
	Version            string
	GitHash            string
	DatabaseURL        string
	LLMBaseURL         string
	LLMAPIKey          string
	LLMCompletionModel string
	LLMEmbeddingModel  string
	SMTPHost           string
	SMTPPort           int
	SMTPUsername       string
	SMTPPassword       string
	SMTPFromEmail      string
	SMTPFromName       string
	SMTPEncryption     string
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return Config{
		AppName:            getEnv("APP_NAME", "insurance-core-api"),
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		Version:            getEnv("APP_VERSION", Version),
		GitHash:            getEnv("GIT_HASH", GitHash),
		DatabaseURL:        databaseURL,
		LLMBaseURL:         os.Getenv("LLM_BASE_URL"),
		LLMAPIKey:          os.Getenv("LLM_API_KEY"),
		LLMCompletionModel: os.Getenv("LLM_COMPLETION_MODEL"),
		LLMEmbeddingModel:  os.Getenv("LLM_EMBEDDING_MODEL"),
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           getEnvInt("SMTP_PORT", 587),
		SMTPUsername:       os.Getenv("SMTP_USERNAME"),
		SMTPPassword:       os.Getenv("SMTP_PASSWORD"),
		SMTPFromEmail:      os.Getenv("SMTP_FROM_EMAIL"),
		SMTPFromName:       getEnv("SMTP_FROM_NAME", "Insurance Core"),
		SMTPEncryption:     strings.ToLower(getEnv("SMTP_ENCRYPTION", "starttls")),
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
