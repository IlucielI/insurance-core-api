package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/redis/go-redis/v9"
)

const defaultRedisTimeout = 5 * time.Second

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Timeout  time.Duration
}

type Client struct {
	addr     string
	password string
	db       int
	timeout  time.Duration

	mu     sync.Mutex
	client *redis.Client

	connectFunc func() (*redis.Client, error)
	getFunc     func(ctx context.Context, key string) (string, error)
	setFunc     func(ctx context.Context, key string, value any, expiration time.Duration) error
	delFunc     func(ctx context.Context, keys ...string) error
	existsFunc  func(ctx context.Context, keys ...string) (int64, error)
	pingFunc    func(ctx context.Context) error
	closeFunc   func() error
}

var _ ports.Cache = (*Client)(nil)

func NewClient(config Config) (*Client, error) {
	addr, password, db, err := parseRedisConfig(config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		addr:     addr,
		password: password,
		db:       db,
		timeout:  normalizeRedisTimeout(config.Timeout),
	}

	client.connectFunc = func() (*redis.Client, error) {
		rdb := redis.NewClient(&redis.Options{
			Addr:         client.addr,
			Password:     client.password,
			DB:           client.db,
			DialTimeout:  client.timeout,
			ReadTimeout:  client.timeout,
			WriteTimeout: client.timeout,
		})
		return rdb, nil
	}

	client.getFunc = func(ctx context.Context, key string) (string, error) {
		rdb, err := client.ensureClient()
		if err != nil {
			return "", err
		}
		val, err := rdb.Get(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			return "", ports.ErrCacheMiss
		}
		return val, err
	}

	client.setFunc = func(ctx context.Context, key string, value any, expiration time.Duration) error {
		rdb, err := client.ensureClient()
		if err != nil {
			return err
		}
		return rdb.Set(ctx, key, value, expiration).Err()
	}

	client.delFunc = func(ctx context.Context, keys ...string) error {
		rdb, err := client.ensureClient()
		if err != nil {
			return err
		}
		return rdb.Del(ctx, keys...).Err()
	}

	client.existsFunc = func(ctx context.Context, keys ...string) (int64, error) {
		rdb, err := client.ensureClient()
		if err != nil {
			return 0, err
		}
		return rdb.Exists(ctx, keys...).Result()
	}

	client.pingFunc = func(ctx context.Context) error {
		rdb, err := client.ensureClient()
		if err != nil {
			return err
		}
		return rdb.Ping(ctx).Err()
	}

	client.closeFunc = func() error {
		if client.client == nil {
			return nil
		}
		return client.client.Close()
	}

	return client, nil
}

func (client *Client) Close() error {
	if client == nil {
		return nil
	}
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.closeFunc != nil {
		if err := client.closeFunc(); err != nil {
			return err
		}
	} else if client.client != nil {
		if err := client.client.Close(); err != nil {
			return err
		}
	}
	client.client = nil
	return nil
}

func (client *Client) Get(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return "", errors.New("key is required")
	}

	return client.getFunc(ctx, key)
}

func (client *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("key is required")
	}

	return client.setFunc(ctx, key, value, expiration)
}

func (client *Client) GetJSON(ctx context.Context, key string, dest any) error {
	if dest == nil {
		return errors.New("destination cannot be nil")
	}

	val, err := client.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (client *Client) SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	return client.Set(ctx, key, bytes, expiration)
}

func (client *Client) Delete(ctx context.Context, keys ...string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	validKeys := filterNonEmpty(keys)
	if len(validKeys) == 0 {
		return nil
	}

	return client.delFunc(ctx, validKeys...)
}

func (client *Client) Exists(ctx context.Context, keys ...string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	validKeys := filterNonEmpty(keys)
	if len(validKeys) == 0 {
		return false, nil
	}

	count, err := client.existsFunc(ctx, validKeys...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (client *Client) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return client.pingFunc(ctx)
}

func (client *Client) ensureClient() (*redis.Client, error) {
	if client == nil {
		return nil, errors.New("redis client is required")
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	if client.client != nil {
		return client.client, nil
	}

	if client.connectFunc == nil {
		client.connectFunc = func() (*redis.Client, error) {
			rdb := redis.NewClient(&redis.Options{
				Addr:         client.addr,
				Password:     client.password,
				DB:           client.db,
				DialTimeout:  client.timeout,
				ReadTimeout:  client.timeout,
				WriteTimeout: client.timeout,
			})
			return rdb, nil
		}
	}

	rdb, err := client.connectFunc()
	if err != nil {
		return nil, fmt.Errorf("connect redis: %w", err)
	}
	client.client = rdb
	return client.client, nil
}

func parseRedisConfig(config Config) (addr string, password string, db int, err error) {
	host := strings.TrimSpace(config.Host)
	if host == "" {
		return "", "", 0, errors.New("host is required")
	}

	password = config.Password
	db = config.DB

	if strings.HasPrefix(host, "redis://") || strings.HasPrefix(host, "rediss://") {
		opt, err := redis.ParseURL(host)
		if err != nil {
			return "", "", 0, fmt.Errorf("parse redis url: %w", err)
		}
		if opt.Password != "" {
			password = opt.Password
		}
		db = opt.DB
		return opt.Addr, password, db, nil
	}

	if strings.Contains(host, "://") {
		parsed, err := url.Parse(host)
		if err != nil {
			return "", "", 0, errors.New("invalid url format")
		}
		return "", "", 0, fmt.Errorf("unsupported redis scheme %q", parsed.Scheme)
	}

	if strings.Contains(host, ":") {
		return host, password, db, nil
	}

	port := config.Port
	if port <= 0 {
		port = 6379
	}
	return fmt.Sprintf("%s:%d", host, port), password, db, nil
}

func normalizeRedisTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultRedisTimeout
	}
	return timeout
}

func filterNonEmpty(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
