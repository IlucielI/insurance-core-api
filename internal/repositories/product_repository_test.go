package repositories

import (
	"context"
	"encoding/json"
	"strings"
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

func (c *fakeCache) DeletePrefix(ctx context.Context, prefix string) error {
	for k := range c.store {
		if strings.HasPrefix(k, prefix) {
			delete(c.store, k)
		}
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
	if filterKey != "tenant:global:catalog:products:list:life:active:all::10:0" {
		t.Fatalf("filterKey = %q, want tenant:global:catalog:products:list:life:active:all::10:0", filterKey)
	}

	filterPage2 := ProductFilter{Category: "life", Limit: 10, Offset: 10}
	filterKeyPage2 := productFilterCacheKey(defaultTenantScope, filterPage2)
	if filterKeyPage2 != "tenant:global:catalog:products:list:life:active:all::10:10" {
		t.Fatalf("filterKeyPage2 = %q, want tenant:global:catalog:products:list:life:active:all::10:10", filterKeyPage2)
	}
	if filterKey == filterKeyPage2 {
		t.Fatal("cache key collision between page 1 and page 2")
	}

	// Filter with search
	filterWithSearch := ProductFilter{Category: "life", Search: "secure", Limit: 10, Offset: 0}
	searchKey := productFilterCacheKey(defaultTenantScope, filterWithSearch)
	if searchKey != "tenant:global:catalog:products:list:life:active:all:secure:10:0" {
		t.Fatalf("searchKey = %q, want tenant:global:catalog:products:list:life:active:all:secure:10:0", searchKey)
	}
	if filterKey == searchKey {
		t.Fatal("cache key collision between unfiltered and search query")
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
	filterWithColon := ProductFilter{Category: "life:extra", Search: "term:with:colon", Limit: 10, Offset: 0}
	sanitizedKey := productFilterCacheKey(defaultTenantScope, filterWithColon)
	if sanitizedKey != "tenant:global:catalog:products:list:life_extra:active:all:term_with_colon:10:0" {
		t.Fatalf("sanitizedKey = %q, want tenant:global:catalog:products:list:life_extra:active:all:term_with_colon:10:0", sanitizedKey)
	}
}

func TestProductManagementRepositoryCRUD(t *testing.T) {
	db := sqliteDB(t)
	cache := newFakeCache()
	repo := NewPostgresProductRepository(db).WithCache(cache)
	pricingRepo := NewPostgresPricingRuleRepository(db).WithCache(cache)

	ctx := context.Background()

	// 1. Create
	newProd := models.Product{
		ID:               "prod-test-1",
		Name:             "Jiwa Sejahtera",
		Slug:             "jiwa-sejahtera",
		Category:         models.ProductCategoryLife,
		Status:           models.ProductStatusDraft,
		ShortDescription: "Asuransi jiwa",
		Description:      "Deskripsi lengkap",
		TargetCustomer:   "Keluarga",
		MinSumAssured:    10000000,
		MaxSumAssured:    500000000,
		MinPaymentTerm:   5,
		MaxPaymentTerm:   20,
		StartingPremium:  100000,
		NonMCULimit:      500000000,
	}
	if err := repo.Create(ctx, &newProd); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 2. FindByID
	found, err := repo.FindByID(ctx, "prod-test-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Name != "Jiwa Sejahtera" || found.Status != models.ProductStatusDraft {
		t.Fatalf("FindByID() got %+v, want Jiwa Sejahtera with draft status", found)
	}

	// 3. Update
	found.Name = "Jiwa Sejahtera Pro"
	if err := repo.Update(ctx, &found); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	updated, err := repo.FindByID(ctx, "prod-test-1")
	if err != nil {
		t.Fatalf("FindByID() after update error = %v", err)
	}
	if updated.Name != "Jiwa Sejahtera Pro" {
		t.Fatalf("updated.Name = %s, want Jiwa Sejahtera Pro", updated.Name)
	}

	// 4. UpdateStatus
	if err := repo.UpdateStatus(ctx, "prod-test-1", models.ProductStatusActive); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	foundActive, err := repo.FindByID(ctx, "prod-test-1")
	if err != nil {
		t.Fatalf("FindByID() after status update error = %v", err)
	}
	if foundActive.Status != models.ProductStatusActive {
		t.Fatalf("foundActive.Status = %s, want active", foundActive.Status)
	}

	// 5. Pricing rules batch save
	rules := []models.ProductPricingRule{
		{RuleCode: "base_rate", RuleName: "Base Rate", RuleType: "base_rate", Factors: map[string]any{"rate": 1.5}, IsActive: true},
	}
	if err := pricingRepo.SaveBatch(ctx, "prod-test-1", rules); err != nil {
		t.Fatalf("SaveBatch() error = %v", err)
	}
	savedRules, err := pricingRepo.FindByProductID(ctx, "prod-test-1")
	if err != nil {
		t.Fatalf("FindByProductID() error = %v", err)
	}
	if len(savedRules) != 1 || savedRules[0].RuleCode != "base_rate" {
		t.Fatalf("savedRules = %+v, want 1 rule with base_rate", savedRules)
	}

	// 6. HasApplications (false initially)
	hasApps, err := repo.HasApplications(ctx, "prod-test-1")
	if err != nil {
		t.Fatalf("HasApplications() error = %v", err)
	}
	if hasApps {
		t.Fatalf("HasApplications() = true, want false")
	}

	// Insert an application
	app := models.Application{
		ID:         "app-1",
		ProductID:  "prod-test-1",
		FullName:   "Budi",
		Email:      "budi@example.com",
		Phone:      "08123456789",
		Premium:    500000,
		Status:     models.ApplicationStatusApproved,
	}
	if err := db.Create(&app).Error; err != nil {
		t.Fatalf("db.Create(app) error = %v", err)
	}

	hasApps, err = repo.HasApplications(ctx, "prod-test-1")
	if err != nil {
		t.Fatalf("HasApplications() after create error = %v", err)
	}
	if !hasApps {
		t.Fatalf("HasApplications() = false, want true")
	}

	// 7. GetMetrics
	metrics, err := repo.GetMetrics(ctx)
	if err != nil {
		t.Fatalf("GetMetrics() error = %v", err)
	}
	if metrics.TotalProducts < 1 {
		t.Fatalf("metrics.TotalProducts = %d, want >= 1", metrics.TotalProducts)
	}
	if metrics.TotalActivePolicies != 1 {
		t.Fatalf("metrics.TotalActivePolicies = %d, want 1", metrics.TotalActivePolicies)
	}
	if metrics.TotalGWPVolume != 500000 {
		t.Fatalf("metrics.TotalGWPVolume = %d, want 500000", metrics.TotalGWPVolume)
	}

	// 8. Delete
	if err := repo.Delete(ctx, "prod-test-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	_, err = repo.FindByID(ctx, "prod-test-1")
	if err == nil {
		t.Fatalf("FindByID() after delete should return error")
	}
}

func TestProductRepositoryCacheInvalidation(t *testing.T) {
	db := sqliteDB(t)
	cache := newFakeCache()
	repo := NewPostgresProductRepository(db).WithCache(cache)
	ctx := context.Background()

	// 1. Initial product
	p := models.Product{
		ID:               "prod-cache-inv",
		Name:             "Cache Invalidation Test",
		Slug:             "cache-inv-test",
		Category:         models.ProductCategoryLife,
		Status:           models.ProductStatusActive,
		ShortDescription: "short",
		Description:      "long",
		TargetCustomer:   "all",
		MinSumAssured:    10000000,
		MaxSumAssured:    100000000,
		MinPaymentTerm:   5,
		MaxPaymentTerm:   20,
		StartingPremium:  100000,
	}
	if err := repo.Create(ctx, &p); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// 2. Populate cache: FindAll and FindBySlug and FindByID
	filter := ProductFilter{Category: "life", Limit: 10}
	_, err := repo.FindAll(ctx, filter)
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	_, err = repo.FindBySlug(ctx, "cache-inv-test")
	if err != nil {
		t.Fatalf("FindBySlug() error = %v", err)
	}
	_, err = repo.FindByID(ctx, "prod-cache-inv")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	// Check that keys exist in cache
	listKey := productFilterCacheKey(defaultTenantScope, filter)
	slugKey := "tenant:global:catalog:products:slug:cache-inv-test"
	idKey := "tenant:global:catalog:products:id:prod-cache-inv"

	if _, ok := cache.store[listKey]; !ok {
		t.Fatalf("expected list key %q in cache", listKey)
	}
	if _, ok := cache.store[slugKey]; !ok {
		t.Fatalf("expected slug key %q in cache", slugKey)
	}
	if _, ok := cache.store[idKey]; !ok {
		t.Fatalf("expected id key %q in cache", idKey)
	}

	// 3. Mutate: Update product status
	if err := repo.UpdateStatus(ctx, "prod-cache-inv", models.ProductStatusDraft); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	// Invalidation check: list key and item key must have been deleted
	if _, ok := cache.store[listKey]; ok {
		t.Fatalf("list key %q was NOT invalidated after UpdateStatus", listKey)
	}
	if _, ok := cache.store[idKey]; ok {
		t.Fatalf("id key %q was NOT invalidated after UpdateStatus", idKey)
	}
}


