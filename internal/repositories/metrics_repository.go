package repositories

import (
	"context"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

type MetricsRepository interface {
	GetAdminMetrics(ctx context.Context) (models.AdminMetrics, error)
}

type PostgresMetricsRepository struct {
	db *gorm.DB
}

func NewPostgresMetricsRepository(db *gorm.DB) *PostgresMetricsRepository {
	return &PostgresMetricsRepository{db: db}
}

func (repository *PostgresMetricsRepository) GetAdminMetrics(ctx context.Context) (models.AdminMetrics, error) {
	var metrics models.AdminMetrics

	// 1. Status Counts
	var statusCounts []models.ApplicationStatusCount
	if err := repository.db.WithContext(ctx).
		Model(&models.Application{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return metrics, err
	}
	metrics.StatusCounts = statusCounts

	// 2. Active Policies & Written Premium (status = approved)
	type premiumAgg struct {
		TotalPremium int64
		Count        int64
	}
	var pAgg premiumAgg
	if err := repository.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("status = ?", models.ApplicationStatusApproved).
		Select("COALESCE(SUM(premium), 0) as total_premium, count(*) as count").
		Scan(&pAgg).Error; err != nil {
		return metrics, err
	}
	metrics.TotalWrittenPremium = pAgg.TotalPremium
	metrics.ActivePoliciesCount = pAgg.Count

	// 3. SLA Review times
	type reviewTimes struct {
		CreatedAt  time.Time
		ReviewedAt *time.Time
	}
	var reviewedApps []reviewTimes
	if err := repository.db.WithContext(ctx).
		Model(&models.Application{}).
		Where("reviewed_at IS NOT NULL").
		Select("created_at, reviewed_at").
		Scan(&reviewedApps).Error; err != nil {
		return metrics, err
	}

	var totalDuration float64
	var totalWithinSLA int64
	for _, app := range reviewedApps {
		if app.ReviewedAt != nil {
			durMinutes := app.ReviewedAt.Sub(app.CreatedAt).Minutes()
			if durMinutes < 0 {
				durMinutes = 0
			}
			totalDuration += durMinutes
			if durMinutes <= 60 {
				totalWithinSLA++
			}
		}
	}
	metrics.SLAMetrics = models.SLAMetricsAggregate{
		TotalReviewed:     int64(len(reviewedApps)),
		TotalWithinSLA:    totalWithinSLA,
		TotalDurationMins: totalDuration,
	}

	// 4. Top Products by Volume
	var topProds []models.ProductVolumeAggregate
	if err := repository.db.WithContext(ctx).
		Table("applications").
		Select("applications.product_id, products.slug as product_slug, products.name as product_name, products.category, count(applications.id) as active_policies_count, COALESCE(sum(applications.premium), 0) as total_premium").
		Joins("JOIN products ON products.id = applications.product_id").
		Where("applications.status = ?", models.ApplicationStatusApproved).
		Group("applications.product_id, products.slug, products.name, products.category").
		Order("total_premium DESC").
		Limit(5).
		Scan(&topProds).Error; err != nil {
		return metrics, err
	}
	metrics.TopProducts = topProds

	return metrics, nil
}
