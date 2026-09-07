package ports

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss indicates that the requested key does not exist in cache.
var ErrCacheMiss = errors.New("cache: key not found")

// Cache defines standard key-value caching operations.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	GetJSON(ctx context.Context, key string, dest any) error
	SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (bool, error)
	Ping(ctx context.Context) error
	Close() error
}
