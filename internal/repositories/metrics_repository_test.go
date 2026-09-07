package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

func TestPostgresMetricsRepository(t *testing.T) {
	db := sqliteDB(t)
	seedProduct(t, db, models.Product{ID: "prod-1", Name: "Secure Life Plus", Slug: "secure-life-plus", Category: models.ProductCategoryLife})
	seedProduct(t, db, models.Product{ID: "prod-2", Name: "Health Guard", Slug: "health-guard", Category: models.ProductCategoryHealth})

	now := time.Now()
	reviewedAt := now.Add(20 * time.Minute)

	apps := []models.Application{
		{
			ID:          "app-1",
			ProductID:   "prod-1",
			FullName:    "Budi Santoso",
			Email:       "budi@example.com",
			Phone:       "08123456789",
			Age:         30,
			Gender:      "male",
			SumAssured:  500000000,
			PaymentTerm: 10,
			Premium:     1500000,
			Status:      models.ApplicationStatusApproved,
			CreatedAt:   now,
			ReviewedAt:  &reviewedAt,
		},
		{
			ID:          "app-2",
			ProductID:   "prod-1",
			FullName:    "Siti Rahma",
			Email:       "siti@example.com",
			Phone:       "08123456780",
			Age:         28,
			Gender:      "female",
			SumAssured:  300000000,
			PaymentTerm: 5,
			Premium:     1000000,
			Status:      models.ApplicationStatusUnderReview,
			CreatedAt:   now,
		},
		{
			ID:          "app-3",
			ProductID:   "prod-2",
			FullName:    "Agus Wijaya",
			Email:       "agus@example.com",
			Phone:       "08123456781",
			Age:         35,
			Gender:      "male",
			SumAssured:  200000000,
			PaymentTerm: 10,
			Premium:     800000,
			Status:      models.ApplicationStatusRejected,
			CreatedAt:   now,
			ReviewedAt:  &reviewedAt,
		},
	}

	for _, a := range apps {
		if err := db.Create(&a).Error; err != nil {
			t.Fatalf("Failed to seed application: %v", err)
		}
	}

	repo := NewPostgresMetricsRepository(db)
	metrics, err := repo.GetAdminMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetAdminMetrics error: %v", err)
	}

	if metrics.ActivePoliciesCount != 1 {
		t.Errorf("ActivePoliciesCount = %d, want 1", metrics.ActivePoliciesCount)
	}
	if metrics.TotalWrittenPremium != 1500000 {
		t.Errorf("TotalWrittenPremium = %d, want 1500000", metrics.TotalWrittenPremium)
	}
	if metrics.SLAMetrics.TotalReviewed != 2 {
		t.Errorf("TotalReviewed = %d, want 2", metrics.SLAMetrics.TotalReviewed)
	}
	if metrics.SLAMetrics.TotalWithinSLA != 2 {
		t.Errorf("TotalWithinSLA = %d, want 2", metrics.SLAMetrics.TotalWithinSLA)
	}
	if len(metrics.TopProducts) != 1 {
		t.Errorf("TopProducts count = %d, want 1", len(metrics.TopProducts))
	} else {
		if metrics.TopProducts[0].ProductSlug != "secure-life-plus" {
			t.Errorf("TopProducts[0].ProductSlug = %s, want secure-life-plus", metrics.TopProducts[0].ProductSlug)
		}
	}
}
