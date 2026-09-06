package redis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/ports"
)

func TestNewClientValidatesConfig(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatal("NewClient(Config{}) error = nil, want error")
	}

	if _, err := NewClient(Config{Host: "ftp://localhost"}); err == nil {
		t.Fatal("NewClient(ftp://) error = nil, want error")
	}

	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient(localhost) error = %v", err)
	}
	if client.addr != "localhost:6379" {
		t.Fatalf("client.addr = %s, want localhost:6379", client.addr)
	}
}

func TestParseRedisConfig(t *testing.T) {
	testAuth := os.Getenv("TEST_REDIS_PASSWORD")
	if testAuth == "" {
		testAuth = "auth-" + t.Name()
	}
	redisURLWithAuth := fmt.Sprintf("redis://:%s@localhost:6381/2", testAuth)

	tests := []struct {
		name         string
		config       Config
		wantAddr     string
		wantPassword string
		wantDB       int
		wantErr      bool
	}{
		{
			name:     "plain host with default port",
			config:   Config{Host: "127.0.0.1"},
			wantAddr: "127.0.0.1:6379",
			wantDB:   0,
		},
		{
			name:         "host with custom port",
			config:       Config{Host: "127.0.0.1", Port: 6380, Password: testAuth, DB: 1},
			wantAddr:     "127.0.0.1:6380",
			wantPassword: testAuth,
			wantDB:       1,
		},
		{
			name:     "host:port string",
			config:   Config{Host: "redis.internal:6379"},
			wantAddr: "redis.internal:6379",
		},
		{
			name:         "redis url with auth and db",
			config:       Config{Host: redisURLWithAuth},
			wantAddr:     "localhost:6381",
			wantPassword: testAuth,
			wantDB:       2,
		},
		{
			name:    "empty host",
			config:  Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, password, db, err := parseRedisConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseRedisConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if addr != tt.wantAddr {
				t.Errorf("addr = %s, want %s", addr, tt.wantAddr)
			}
			if password != tt.wantPassword {
				t.Errorf("password = %s, want %s", password, tt.wantPassword)
			}
			if db != tt.wantDB {
				t.Errorf("db = %d, want %d", db, tt.wantDB)
			}
		})
	}
}

func TestGetAndSet(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	storage := make(map[string]string)
	client.setFunc = func(ctx context.Context, key string, value any, expiration time.Duration) error {
		storage[key] = value.(string)
		return nil
	}
	client.getFunc = func(ctx context.Context, key string) (string, error) {
		val, ok := storage[key]
		if !ok {
			return "", ports.ErrCacheMiss
		}
		return val, nil
	}

	if err := client.Set(context.Background(), "user:1", "John", time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	val, err := client.Get(context.Background(), "user:1")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if val != "John" {
		t.Fatalf("Get() = %q, want John", val)
	}

	// Key validation
	if err := client.Set(context.Background(), " ", "val", 0); err == nil {
		t.Fatal("Set(empty key) error = nil, want error")
	}
	if _, err := client.Get(context.Background(), ""); err == nil {
		t.Fatal("Get(empty key) error = nil, want error")
	}
}

func TestGetNotFoundReturnsCacheMiss(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	client.getFunc = func(ctx context.Context, key string) (string, error) {
		return "", ports.ErrCacheMiss
	}

	_, err = client.Get(context.Background(), "missing")
	if !errors.Is(err, ports.ErrCacheMiss) {
		t.Fatalf("Get() error = %v, want ErrCacheMiss", err)
	}
}

func TestGetJSONAndSetJSON(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	type Profile struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var stored string
	client.setFunc = func(ctx context.Context, key string, value any, expiration time.Duration) error {
		stored = string(value.([]byte))
		return nil
	}
	client.getFunc = func(ctx context.Context, key string) (string, error) {
		return stored, nil
	}

	expected := Profile{Name: "Alice", Age: 30}
	if err := client.SetJSON(context.Background(), "profile:1", expected, time.Minute); err != nil {
		t.Fatalf("SetJSON() error = %v", err)
	}

	var actual Profile
	if err := client.GetJSON(context.Background(), "profile:1", &actual); err != nil {
		t.Fatalf("GetJSON() error = %v", err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("GetJSON() = %+v, want %+v", actual, expected)
	}

	if err := client.GetJSON(context.Background(), "profile:1", nil); err == nil {
		t.Fatal("GetJSON(nil dest) error = nil, want error")
	}
}

func TestDeleteAndExists(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var deleted []string
	client.delFunc = func(ctx context.Context, keys ...string) error {
		deleted = keys
		return nil
	}

	client.existsFunc = func(ctx context.Context, keys ...string) (int64, error) {
		if keys[0] == "present" {
			return 1, nil
		}
		return 0, nil
	}

	// Delete empty
	if err := client.Delete(context.Background()); err != nil {
		t.Fatalf("Delete() empty error = %v", err)
	}

	// Delete with keys
	if err := client.Delete(context.Background(), "k1", "k2"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if len(deleted) != 2 || deleted[0] != "k1" || deleted[1] != "k2" {
		t.Fatalf("deleted = %+v", deleted)
	}

	// Exists empty
	exists, err := client.Exists(context.Background())
	if err != nil || exists {
		t.Fatalf("Exists() empty = (%v, %v), want (false, nil)", exists, err)
	}

	// Exists present
	exists, err = client.Exists(context.Background(), "present")
	if err != nil || !exists {
		t.Fatalf("Exists(present) = (%v, %v), want (true, nil)", exists, err)
	}

	// Exists absent
	exists, err = client.Exists(context.Background(), "absent")
	if err != nil || exists {
		t.Fatalf("Exists(absent) = (%v, %v), want (false, nil)", exists, err)
	}
}

func TestPing(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var pingCalled bool
	client.pingFunc = func(ctx context.Context) error {
		pingCalled = true
		return nil
	}

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}
	if !pingCalled {
		t.Fatal("pingFunc was not called")
	}
}

func TestClose(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	var closeCalled bool
	client.closeFunc = func() error {
		closeCalled = true
		return nil
	}

	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !closeCalled {
		t.Fatal("closeFunc was not called")
	}

	// Close on nil client
	var nilClient *Client
	if err := nilClient.Close(); err != nil {
		t.Fatalf("nilClient.Close() error = %v", err)
	}
}

func TestContextCancelled(t *testing.T) {
	client, err := NewClient(Config{Host: "localhost"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Get(ctx, "k"); err == nil {
		t.Fatal("Get() with cancelled ctx error = nil, want error")
	}
	if err := client.Set(ctx, "k", "v", 0); err == nil {
		t.Fatal("Set() with cancelled ctx error = nil, want error")
	}
	if err := client.Delete(ctx, "k"); err == nil {
		t.Fatal("Delete() with cancelled ctx error = nil, want error")
	}
	if _, err := client.Exists(ctx, "k"); err == nil {
		t.Fatal("Exists() with cancelled ctx error = nil, want error")
	}
	if err := client.Ping(ctx); err == nil {
		t.Fatal("Ping() with cancelled ctx error = nil, want error")
	}
}
