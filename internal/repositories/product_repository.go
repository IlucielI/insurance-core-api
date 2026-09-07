package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New(constants.ErrProductNotFound)

type ProductRepository interface {
	FindAll(ctx context.Context, filter ProductFilter) ([]models.Product, error)
	FindByID(ctx context.Context, id string) (models.Product, error)
	FindBySlug(ctx context.Context, slug string) (models.Product, error)
	Create(ctx context.Context, product *models.Product) error
	Update(ctx context.Context, product *models.Product) error
	UpdateStatus(ctx context.Context, id string, status models.ProductStatus) error
	Delete(ctx context.Context, id string) error
	GetMetrics(ctx context.Context) (dtos.ProductManagementMetricsResponse, error)
	HasApplications(ctx context.Context, productID string) (bool, error)
}

type ProductFilter struct {
	Category   string
	Status     string
	IsFeatured *bool
	Limit      int
	Offset     int
	Search     string
}

type PostgresProductRepository struct {
	db    *gorm.DB
	cache ports.Cache
}

func NewPostgresProductRepository(db *gorm.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (repository *PostgresProductRepository) WithCache(cache ports.Cache) *PostgresProductRepository {
	repository.cache = cache
	return repository
}

const defaultTenantScope = "global"

func (repository *PostgresProductRepository) FindAll(ctx context.Context, filter ProductFilter) ([]models.Product, error) {
	cacheKey := productFilterCacheKey(defaultTenantScope, filter)
	if repository.cache != nil && cacheKey != "" {
		var cached []models.Product
		if err := repository.cache.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	query := repository.db.WithContext(ctx).Order("created_at ASC")

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	status := strings.TrimSpace(filter.Status)
	if status == "" {
		status = string(models.ProductStatusActive)
	}
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	if filter.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filter.IsFeatured)
	}

	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(short_description) LIKE ?", searchTerm, searchTerm)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	if repository.cache != nil && cacheKey != "" && len(products) > 0 {
		if err := repository.cache.SetJSON(ctx, cacheKey, products, 15*time.Minute); err != nil {
			log.Printf("[ProductRepository] failed to cache products list: %v", err)
		}
	}

	return products, nil
}

func (repository *PostgresProductRepository) FindByID(ctx context.Context, id string) (models.Product, error) {
	cacheKey := fmt.Sprintf("tenant:%s:catalog:products:id:%s", defaultTenantScope, sanitizeCacheSegment(id))
	if repository.cache != nil {
		var cachedProduct models.Product
		if err := repository.cache.GetJSON(ctx, cacheKey, &cachedProduct); err == nil && cachedProduct.ID != "" {
			return cachedProduct, nil
		}
	}

	var product models.Product
	err := repository.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Product{}, ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, err
	}

	if repository.cache != nil {
		if err := repository.cache.SetJSON(ctx, cacheKey, product, 1*time.Hour); err != nil {
			log.Printf("[ProductRepository] warning: failed to cache product %s: %v", id, err)
		}
	}

	return product, nil
}

func (repository *PostgresProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	cacheKey := fmt.Sprintf("tenant:%s:catalog:products:slug:%s", defaultTenantScope, sanitizeCacheSegment(slug))
	if repository.cache != nil {
		var cachedProduct models.Product
		if err := repository.cache.GetJSON(ctx, cacheKey, &cachedProduct); err == nil && cachedProduct.ID != "" {
			return cachedProduct, nil
		}
	}

	var product models.Product
	err := repository.db.WithContext(ctx).Where("slug = ? OR id = ?", slug, slug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Product{}, ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, err
	}

	if repository.cache != nil {
		if err := repository.cache.SetJSON(ctx, cacheKey, product, 1*time.Hour); err != nil {
			log.Printf("[ProductRepository] failed to cache product slug %s: %v", slug, err)
		}
	}

	return product, nil
}

func (repository *PostgresProductRepository) Create(ctx context.Context, product *models.Product) error {
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	if product.Status == "" {
		product.Status = models.ProductStatusActive
	}
	now := time.Now().UTC()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	product.UpdatedAt = now

	if err := repository.db.WithContext(ctx).Create(product).Error; err != nil {
		return err
	}

	repository.invalidateProductCache(ctx, product.ID, product.Slug)

	return nil
}

func (repository *PostgresProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now().UTC()
	result := repository.db.WithContext(ctx).Save(product)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	repository.invalidateProductCache(ctx, product.ID, product.Slug)

	return nil
}

func (repository *PostgresProductRepository) UpdateStatus(ctx context.Context, id string, status models.ProductStatus) error {
	result := repository.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ? OR slug = ?", id, id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProductNotFound
	}

	repository.invalidateProductCache(ctx, id, id)

	return nil
}

func (repository *PostgresProductRepository) Delete(ctx context.Context, id string) error {
	var product models.Product
	err := repository.db.WithContext(ctx).Where("id = ? OR slug = ?", id, id).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrProductNotFound
	}
	if err != nil {
		return err
	}

	result := repository.db.WithContext(ctx).Where("id = ?", product.ID).Delete(&models.Product{})
	if result.Error != nil {
		return result.Error
	}

	repository.invalidateProductCache(ctx, product.ID, product.Slug)

	return nil
}

func (repository *PostgresProductRepository) invalidateProductCache(ctx context.Context, id, slug string) {
	if repository.cache == nil {
		return
	}
	keys := make([]string, 0, 2)
	if slug != "" {
		keys = append(keys, fmt.Sprintf("tenant:%s:catalog:products:slug:%s", defaultTenantScope, sanitizeCacheSegment(slug)))
	}
	if id != "" {
		keys = append(keys, fmt.Sprintf("tenant:%s:catalog:products:id:%s", defaultTenantScope, sanitizeCacheSegment(id)))
	}
	if len(keys) > 0 {
		if err := repository.cache.Delete(ctx, keys...); err != nil {
			log.Printf("[ProductRepository] warning: failed to delete product cache keys: %v", err)
		}
	}
	if err := repository.cache.DeletePrefix(ctx, fmt.Sprintf("tenant:%s:catalog:products:list:", defaultTenantScope)); err != nil {
		log.Printf("[ProductRepository] warning: failed to delete prefix catalog cache: %v", err)
	}
}

func (repository *PostgresProductRepository) HasApplications(ctx context.Context, productID string) (bool, error) {
	var count int64
	err := repository.db.WithContext(ctx).Model(&models.Application{}).
		Where("product_id = ?", productID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (repository *PostgresProductRepository) GetMetrics(ctx context.Context) (dtos.ProductManagementMetricsResponse, error) {
	var metrics dtos.ProductManagementMetricsResponse

	type statusCount struct {
		Status string
		Count  int
	}
	var counts []statusCount
	err := repository.db.WithContext(ctx).
		Model(&models.Product{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&counts).Error
	if err != nil {
		return metrics, err
	}

	for _, sc := range counts {
		metrics.TotalProducts += sc.Count
		switch models.ProductStatus(sc.Status) {
		case models.ProductStatusActive:
			metrics.ActiveProducts = sc.Count
		case models.ProductStatusDraft:
			metrics.DraftProducts = sc.Count
		case models.ProductStatusArchived:
			metrics.ArchivedProducts = sc.Count
		}
	}

	type appSummary struct {
		ActivePoliciesCount int   `gorm:"column:active_count"`
		TotalGWP            int64 `gorm:"column:total_gwp"`
	}
	var appSum appSummary
	err = repository.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("COUNT(*) as active_count, COALESCE(SUM(premium), 0) as total_gwp").
		Where("status = ?", models.ApplicationStatusApproved).
		Scan(&appSum).Error
	if err != nil {
		log.Printf("[ProductRepository] warning: failed to query application metrics: %v", err)
	} else {
		metrics.TotalActivePolicies = appSum.ActivePoliciesCount
		metrics.TotalGWPVolume = appSum.TotalGWP
	}

	metrics.TotalGWPVolumeFormatted = dtos.FormatIDR(metrics.TotalGWPVolume)
	metrics.AverageLossRatio = 0.385 // 38.5% industry baseline

	return metrics, nil
}

func productFilterCacheKey(tenant string, filter ProductFilter) string {
	category := sanitizeCacheSegment(filter.Category)
	featured := "all"
	if filter.IsFeatured != nil {
		if *filter.IsFeatured {
			featured = "true"
		} else {
			featured = "false"
		}
	}
	search := sanitizeCacheSegment(filter.Search)
	status := strings.TrimSpace(filter.Status)
	if status == "" {
		status = string(models.ProductStatusActive)
	}
	return fmt.Sprintf("tenant:%s:catalog:products:list:%s:%s:%s:%s:%d:%d", tenant, category, sanitizeCacheSegment(status), featured, search, filter.Limit, filter.Offset)
}

func sanitizeCacheSegment(val string) string {
	return strings.ReplaceAll(strings.TrimSpace(val), ":", "_")
}

