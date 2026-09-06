package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New(constants.ErrProductNotFound)

type ProductRepository interface {
	FindAll(ctx context.Context, filter ProductFilter) ([]models.Product, error)
	FindBySlug(ctx context.Context, slug string) (models.Product, error)
}

type ProductFilter struct {
	Category   string
	IsFeatured *bool
	Limit      int
	Offset     int
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

func (repository *PostgresProductRepository) FindAll(ctx context.Context, filter ProductFilter) ([]models.Product, error) {
	cacheKey := productFilterCacheKey(filter)
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

	if filter.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filter.IsFeatured)
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

func (repository *PostgresProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	cacheKey := "catalog:products:slug:" + slug
	if repository.cache != nil {
		var cachedProduct models.Product
		if err := repository.cache.GetJSON(ctx, cacheKey, &cachedProduct); err == nil && cachedProduct.ID != "" {
			return cachedProduct, nil
		}
	}

	var product models.Product
	err := repository.db.WithContext(ctx).Where("slug = ?", slug).First(&product).Error
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

func productFilterCacheKey(filter ProductFilter) string {
	featured := "all"
	if filter.IsFeatured != nil {
		if *filter.IsFeatured {
			featured = "true"
		} else {
			featured = "false"
		}
	}
	return fmt.Sprintf("catalog:products:list:%s:%s:%d:%d", filter.Category, featured, filter.Limit, filter.Offset)
}
