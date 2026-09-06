package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName             string
	AppEnv              string
	HTTPPort            string
	Version             string
	GitHash             string
	DatabaseURL         string
	LLMBaseURL          string
	LLMAPIKey           string
	LLMCompletionModel  string
	LLMEmbeddingModel   string
	SMTPHost            string
	SMTPPort            int
	SMTPUsername        string
	SMTPPassword        string
	SMTPFromEmail       string
	SMTPFromName        string
	SMTPEncryption      string
	NATSHost            string
	NATSPort            int
	NATSToken           string
	NATSName            string
	NATSTimeout         int
	S3Endpoint          string
	S3AccessKey         string
	S3SecretKey         string
	S3Region            string
	S3Bucket            string
	S3UseSSL            bool
	S3ForcePathStyle    bool
	S3UploadUrlLifetime int
	S3DownloadUrlLifetime int
	S3OverrideBaseURL    string
	RedisHost           string
	RedisPort           int
	RedisPassword       string
	RedisDB             int
	RedisTimeout        int // in seconds
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
		NATSHost:            os.Getenv("NATS_HOST"),
		NATSPort:            getEnvInt("NATS_PORT", 4222),
		NATSToken:           os.Getenv("NATS_TOKEN"),
		NATSName:            getEnv("NATS_NAME", "insurance-core-api"),
		NATSTimeout:         getEnvInt("NATS_TIMEOUT", 5),
		S3Endpoint:          os.Getenv("S3_ENDPOINT"),
		S3AccessKey:         os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:         os.Getenv("S3_SECRET_KEY"),
		S3Region:            getEnv("S3_REGION", "ap-southeast-1"),
		S3Bucket:            os.Getenv("S3_BUCKET"),
		S3UseSSL:            getEnvBool("S3_USE_SSL", false),
		S3ForcePathStyle:    getEnvBool("S3_FORCE_PATH_STYLE", true),
		S3UploadUrlLifetime: getEnvInt("S3_UPLOAD_URL_LIFETIME", 15),
		S3DownloadUrlLifetime: getEnvInt("S3_DOWNLOAD_URL_LIFETIME", 1440),
		S3OverrideBaseURL:   os.Getenv("S3_OVERRIDE_BASE_URL"),
		RedisHost:           os.Getenv("REDIS_HOST"),
		RedisPort:           getEnvInt("REDIS_PORT", 6379),
		RedisPassword:       os.Getenv("REDIS_PASSWORD"),
		RedisDB:             getEnvInt("REDIS_DB", 0),
		RedisTimeout:        getEnvInt("REDIS_TIMEOUT", 5),
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

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
