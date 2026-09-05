package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/app?sslmode=disable")
	t.Setenv("APP_NAME", "insurance-test")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("APP_VERSION", "1.0.0")
	t.Setenv("GIT_HASH", "abc123")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppName != "insurance-test" || cfg.HTTPPort != "9000" || cfg.Version != "1.0.0" || cfg.GitHash != "abc123" {
		t.Fatalf("Load() = %+v, want env values", cfg)
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

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AppName != "insurance-core-api" || cfg.AppEnv != "development" || cfg.HTTPPort != "8080" || cfg.Version != Version || cfg.GitHash != GitHash {
		t.Fatalf("Load() = %+v, want defaults", cfg)
	}
}
