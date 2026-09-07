package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrPricingRuleNotFound = errors.New("pricing rule not found")
)

type PricingRuleRepository interface {
	FindByProductID(ctx context.Context, productID string) ([]models.ProductPricingRule, error)
	FindByProductSlug(ctx context.Context, slug string) ([]models.ProductPricingRule, error)
	Create(ctx context.Context, rule *models.ProductPricingRule) error
	SaveBatch(ctx context.Context, productID string, rules []models.ProductPricingRule) error
}

type PostgresPricingRuleRepository struct {
	db    *gorm.DB
	cache ports.Cache
}

func NewPostgresPricingRuleRepository(db *gorm.DB, cache ...ports.Cache) *PostgresPricingRuleRepository {
	var c ports.Cache
	if len(cache) > 0 {
		c = cache[0]
	}
	return &PostgresPricingRuleRepository{db: db, cache: c}
}

func (r *PostgresPricingRuleRepository) WithCache(cache ports.Cache) *PostgresPricingRuleRepository {
	r.cache = cache
	return r
}

func (r *PostgresPricingRuleRepository) FindByProductID(ctx context.Context, productID string) ([]models.ProductPricingRule, error) {
	cacheKey := fmt.Sprintf("tenant:%s:pricing_rules:product_id:%s", defaultTenantScope, sanitizeCacheSegment(productID))
	if r.cache != nil {
		var cached []models.ProductPricingRule
		if err := r.cache.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	var rules []models.ProductPricingRule
	if err := r.db.WithContext(ctx).
		Where("product_id = ? AND is_active = ?", productID, true).
		Order("order_index ASC, id ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}

	if r.cache != nil && len(rules) > 0 {
		if err := r.cache.SetJSON(ctx, cacheKey, rules, 10*time.Minute); err != nil {
			log.Printf("[PricingRuleRepo] warning: failed to cache rules for %s: %v", productID, err)
		}
	}

	return rules, nil
}

func (r *PostgresPricingRuleRepository) FindByProductSlug(ctx context.Context, slug string) ([]models.ProductPricingRule, error) {
	cacheKey := fmt.Sprintf("tenant:%s:pricing_rules:product_slug:%s", defaultTenantScope, sanitizeCacheSegment(slug))
	if r.cache != nil {
		var cached []models.ProductPricingRule
		if err := r.cache.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	var rules []models.ProductPricingRule
	if err := r.db.WithContext(ctx).
		Table("product_pricing_rules").
		Select("product_pricing_rules.*").
		Joins("JOIN products ON products.id = product_pricing_rules.product_id").
		Where("products.slug = ? AND product_pricing_rules.is_active = ?", slug, true).
		Order("product_pricing_rules.order_index ASC, product_pricing_rules.id ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}

	if r.cache != nil && len(rules) > 0 {
		if err := r.cache.SetJSON(ctx, cacheKey, rules, 10*time.Minute); err != nil {
			log.Printf("[PricingRuleRepo] warning: failed to cache rules for %s: %v", slug, err)
		}
	}

	return rules, nil
}

func (r *PostgresPricingRuleRepository) Create(ctx context.Context, rule *models.ProductPricingRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *PostgresPricingRuleRepository) SaveBatch(ctx context.Context, productID string, rules []models.ProductPricingRule) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", productID).Delete(&models.ProductPricingRule{}).Error; err != nil {
			return err
		}

		for i := range rules {
			rules[i].ProductID = productID
			if rules[i].ID == "" {
				rules[i].ID = uuid.New().String()
			}
			if rules[i].CreatedAt.IsZero() {
				rules[i].CreatedAt = time.Now().UTC()
			}
			rules[i].UpdatedAt = time.Now().UTC()
			if err := tx.Create(&rules[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.Delete(ctx,
			fmt.Sprintf("tenant:%s:pricing_rules:product_id:%s", defaultTenantScope, sanitizeCacheSegment(productID)),
		); err != nil {
			log.Printf("[PricingRuleRepo] warning: failed to delete cache for product %s: %v", productID, err)
		}
		if err := r.cache.DeletePrefix(ctx, fmt.Sprintf("tenant:%s:pricing_rules:", defaultTenantScope)); err != nil {
			log.Printf("[PricingRuleRepo] warning: failed to delete prefix cache: %v", err)
		}
	}

	return nil
}
