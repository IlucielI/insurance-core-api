package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
)

type fakeCache struct {
	store map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{store: make(map[string]string)}
}

func (c *fakeCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := c.store[key]
	if !ok {
		return "", ports.ErrCacheMiss
	}
	return val, nil
}

func (c *fakeCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	c.store[key] = value.(string)
	return nil
}

func (c *fakeCache) GetJSON(ctx context.Context, key string, dest any) error {
	val, ok := c.store[key]
	if !ok {
		return ports.ErrCacheMiss
	}
	return json.Unmarshal([]byte(val), dest)
}

func (c *fakeCache) SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	c.store[key] = string(b)
	return nil
}

func (c *fakeCache) Delete(ctx context.Context, keys ...string) error {
	for _, k := range keys {
		delete(c.store, k)
	}
	return nil
}

func (c *fakeCache) Exists(ctx context.Context, keys ...string) (bool, error) {
	for _, k := range keys {
		if _, ok := c.store[k]; ok {
			return true, nil
		}
	}
	return false, nil
}

func (c *fakeCache) Ping(ctx context.Context) error {
	return nil
}

func (c *fakeCache) Close() error {
	return nil
}

func TestPostgresProductRepositoryWithCache(t *testing.T) {
	cache := newFakeCache()
	// Preload cache with a cached product
	cachedProduct := models.Product{
		ID:   "cached-prod-1",
		Name: "Cached Product",
		Slug: "cached-slug",
	}
	if err := cache.SetJSON(context.Background(), "tenant:global:catalog:products:slug:cached-slug", cachedProduct, time.Hour); err != nil {
		t.Fatalf("cache.SetJSON error = %v", err)
	}

	// Repository with nil db but with cache
	repo := NewPostgresProductRepository(nil).WithCache(cache)

	// Should hit cache and return without touching db
	prod, err := repo.FindBySlug(context.Background(), "cached-slug")
	if err != nil {
		t.Fatalf("FindBySlug(cached-slug) error = %v", err)
	}
	if prod.ID != "cached-prod-1" {
		t.Fatalf("prod.ID = %q, want cached-prod-1", prod.ID)
	}

	// Preload cache for FindAll
	cachedList := []models.Product{cachedProduct}
	filter := ProductFilter{Category: "life", Limit: 10, Offset: 0}
	filterKey := productFilterCacheKey(defaultTenantScope, filter)
	if filterKey != "tenant:global:catalog:products:list:life:all:10:0" {
		t.Fatalf("filterKey = %q, want tenant:global:catalog:products:list:life:all:10:0", filterKey)
	}

	filterPage2 := ProductFilter{Category: "life", Limit: 10, Offset: 10}
	filterKeyPage2 := productFilterCacheKey(defaultTenantScope, filterPage2)
	if filterKeyPage2 != "tenant:global:catalog:products:list:life:all:10:10" {
		t.Fatalf("filterKeyPage2 = %q, want tenant:global:catalog:products:list:life:all:10:10", filterKeyPage2)
	}
	if filterKey == filterKeyPage2 {
		t.Fatal("cache key collision between page 1 and page 2")
	}

	if err := cache.SetJSON(context.Background(), filterKey, cachedList, 15*time.Minute); err != nil {
		t.Fatalf("cache.SetJSON error = %v", err)
	}

	list, err := repo.FindAll(context.Background(), filter)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(list) != 1 || list[0].ID != "cached-prod-1" {
		t.Fatalf("list = %+v, want cached-prod-1", list)
	}

	// Delimiter sanitization test (cache poisoning prevention)
	filterWithColon := ProductFilter{Category: "life:extra", Limit: 10, Offset: 0}
	sanitizedKey := productFilterCacheKey(defaultTenantScope, filterWithColon)
	if sanitizedKey != "tenant:global:catalog:products:list:life_extra:all:10:0" {
		t.Fatalf("sanitizedKey = %q, want tenant:global:catalog:products:list:life_extra:all:10:0", sanitizedKey)
	}
}
