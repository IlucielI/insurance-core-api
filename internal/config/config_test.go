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
}
