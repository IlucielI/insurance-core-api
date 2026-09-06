package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	t.Setenv("APP_NAME", "insurance-test")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("APP_VERSION", "1.0.0")
	t.Setenv("GIT_HASH", "abc123")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USERNAME", "mailer")
	t.Setenv("SMTP_PASSWORD", "secret")
	t.Setenv("SMTP_FROM_EMAIL", "no-reply@example.com")
	t.Setenv("SMTP_FROM_NAME", "Insurance Test")
	t.Setenv("SMTP_ENCRYPTION", "TLS")
	t.Setenv("NATS_HOST", "nats.example.com")
	t.Setenv("NATS_PORT", "4223")
	t.Setenv("NATS_TOKEN", "token")
	t.Setenv("NATS_NAME", "insurance-test")
	t.Setenv("NATS_TIMEOUT", "7")
	t.Setenv("S3_ENDPOINT", "s3.example.com")
	t.Setenv("S3_ACCESS_KEY", "access")
	t.Setenv("S3_SECRET_KEY", "secret")
	t.Setenv("S3_REGION", "ap-southeast-1")
	t.Setenv("S3_BUCKET", "insurance-files")
	t.Setenv("S3_USE_SSL", "true")
	t.Setenv("S3_FORCE_PATH_STYLE", "false")
	t.Setenv("S3_UPLOAD_URL_LIFETIME", "20")
	t.Setenv("S3_DOWNLOAD_URL_LIFETIME", "1445")
	t.Setenv("S3_OVERRIDE_BASE_URL", "https://cdn.example.com/files")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppName != "insurance-test" || cfg.HTTPPort != "9000" || cfg.Version != "1.0.0" || cfg.GitHash != "abc123" {
		t.Fatalf("Load() = %+v, want env values", cfg)
	}
	if cfg.SMTPHost != "smtp.example.com" || cfg.SMTPPort != 2525 || cfg.SMTPUsername != "mailer" || cfg.SMTPPassword != "secret" || cfg.SMTPFromEmail != "no-reply@example.com" || cfg.SMTPFromName != "Insurance Test" || cfg.SMTPEncryption != "tls" {
		t.Fatalf("Load() SMTP config = %+v, want env values", cfg)
	}
	if cfg.NATSHost != "nats.example.com" || cfg.NATSPort != 4223 || cfg.NATSToken != "token" || cfg.NATSName != "insurance-test" || cfg.NATSTimeout != 7 {
		t.Fatalf("Load() NATS config = %+v, want env values", cfg)
	}
	if cfg.S3Endpoint != "s3.example.com" || cfg.S3AccessKey != "access" || cfg.S3SecretKey != "secret" || cfg.S3Region != "ap-southeast-1" || cfg.S3Bucket != "insurance-files" || !cfg.S3UseSSL || cfg.S3ForcePathStyle {
		t.Fatalf("Load() S3 config = %+v, want env values", cfg)
	}
	if cfg.S3UploadUrlLifetime != 20 || cfg.S3DownloadUrlLifetime != 1445 || cfg.S3OverrideBaseURL != "https://cdn.example.com/files" {
		t.Fatalf("Load() S3 lifetime config = %+v, want env values", cfg)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil || err.Error() != "DATABASE_URL is required" {
		t.Fatalf("Load() error = %v, want DATABASE_URL required", err)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	t.Setenv("APP_NAME", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("APP_VERSION", "")
	t.Setenv("GIT_HASH", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_FROM_NAME", "")
	t.Setenv("SMTP_ENCRYPTION", "")
	t.Setenv("NATS_HOST", "")
	t.Setenv("NATS_PORT", "")
	t.Setenv("NATS_TOKEN", "")
	t.Setenv("NATS_NAME", "")
	t.Setenv("NATS_TIMEOUT", "")
	t.Setenv("S3_ENDPOINT", "")
	t.Setenv("S3_ACCESS_KEY", "")
	t.Setenv("S3_SECRET_KEY", "")
	t.Setenv("S3_REGION", "")
	t.Setenv("S3_BUCKET", "")
	t.Setenv("S3_USE_SSL", "")
	t.Setenv("S3_FORCE_PATH_STYLE", "")
	t.Setenv("S3_UPLOAD_URL_LIFETIME", "")
	t.Setenv("S3_DOWNLOAD_URL_LIFETIME", "")
	t.Setenv("S3_OVERRIDE_BASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppName != "insurance-core-api" || cfg.AppEnv != "development" || cfg.HTTPPort != "8080" || cfg.Version != Version || cfg.GitHash != GitHash {
		t.Fatalf("Load() = %+v, want defaults", cfg)
	}
	if cfg.SMTPPort != 587 || cfg.SMTPFromName != "Insurance Core" || cfg.SMTPEncryption != "starttls" {
		t.Fatalf("Load() SMTP defaults = %+v, want defaults", cfg)
	}
	if cfg.NATSHost != "" || cfg.NATSPort != 4222 || cfg.NATSToken != "" || cfg.NATSName != "insurance-core-api" || cfg.NATSTimeout != 5 {
		t.Fatalf("Load() NATS defaults = %+v, want defaults", cfg)
	}
	if cfg.S3Region != "ap-southeast-1" || !cfg.S3ForcePathStyle {
		t.Fatalf("Load() S3 defaults = %+v, want defaults", cfg)
	}
	if cfg.S3UploadUrlLifetime != 15 || cfg.S3DownloadUrlLifetime != 1440 || cfg.S3OverrideBaseURL != "" {
		t.Fatalf("Load() S3 lifetime defaults = %+v, want defaults", cfg)
	}
}
