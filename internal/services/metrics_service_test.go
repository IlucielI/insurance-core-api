package services

import (
	"context"
	"errors"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

type mockMetricsRepository struct {
	data models.AdminMetrics
	err  error
}

func (m *mockMetricsRepository) GetAdminMetrics(ctx context.Context) (models.AdminMetrics, error) {
	return m.data, m.err
}

func TestMetricsService_GetAdminMetrics(t *testing.T) {
	t.Run("success with populated metrics", func(t *testing.T) {
		mockRepo := &mockMetricsRepository{
			data: models.AdminMetrics{
				StatusCounts: []models.ApplicationStatusCount{
					{Status: models.ApplicationStatusApproved, Count: 90},
					{Status: models.ApplicationStatusRejected, Count: 10},
					{Status: models.ApplicationStatusUnderReview, Count: 5},
					{Status: models.ApplicationStatusSubmitted, Count: 5},
				},
				ActivePoliciesCount: 90,
				TotalWrittenPremium: 100000000,
				SLAMetrics: models.SLAMetricsAggregate{
					TotalReviewed:     100,
					TotalWithinSLA:    95,
					TotalDurationMins: 1840.0,
				},
				TopProducts: []models.ProductVolumeAggregate{
					{
						ProductID:           "prod-1",
						ProductSlug:         "secure-life-plus",
						ProductName:         "Secure Life Plus",
						Category:            "life",
						ActivePoliciesCount: 50,
						TotalPremium:        60000000,
					},
				},
			},
		}

		svc := NewMetricsService(mockRepo)
		res, err := svc.GetAdminMetrics(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Applications.Total != 110 {
			t.Errorf("Total apps = %d, want 110", res.Applications.Total)
		}
		if res.Applications.Approved != 90 {
			t.Errorf("Approved = %d, want 90", res.Applications.Approved)
		}
		if res.Applications.ApprovalRate != 90.0 {
			t.Errorf("ApprovalRate = %v, want 90.0", res.Applications.ApprovalRate)
		}
		if res.Premiums.TotalWrittenPremium != 100000000 {
			t.Errorf("TotalWrittenPremium = %d, want 100000000", res.Premiums.TotalWrittenPremium)
		}
		if res.Underwriting.PendingReviewCount != 10 {
			t.Errorf("PendingReviewCount = %d, want 10", res.Underwriting.PendingReviewCount)
		}
		if res.Underwriting.AverageSLAMinutes != 18.4 {
			t.Errorf("AverageSLAMinutes = %v, want 18.4", res.Underwriting.AverageSLAMinutes)
		}
		if res.Underwriting.SLAComplianceRate != 95.0 {
			t.Errorf("SLAComplianceRate = %v, want 95.0", res.Underwriting.SLAComplianceRate)
		}
		if len(res.TopProducts) != 1 {
			t.Fatalf("TopProducts len = %d, want 1", len(res.TopProducts))
		}
		if res.TopProducts[0].Category != "life" {
			t.Errorf("TopProducts category = %s, want life", res.TopProducts[0].Category)
		}
	})

	t.Run("empty database safe defaults", func(t *testing.T) {
		mockRepo := &mockMetricsRepository{
			data: models.AdminMetrics{},
		}

		svc := NewMetricsService(mockRepo)
		res, err := svc.GetAdminMetrics(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res.Applications.Total != 0 || res.Applications.ApprovalRate != 0 {
			t.Errorf("Empty apps total = %d, rate = %v", res.Applications.Total, res.Applications.ApprovalRate)
		}
		if res.Underwriting.SLAComplianceRate != 100.0 {
			t.Errorf("Empty SLA compliance = %v, want 100.0", res.Underwriting.SLAComplianceRate)
		}
		if len(res.TopProducts) != 0 {
			t.Errorf("Empty top products len = %d", len(res.TopProducts))
		}
	})

	t.Run("repository error propagation", func(t *testing.T) {
		mockRepo := &mockMetricsRepository{
			err: errors.New("db connection timeout"),
		}

		svc := NewMetricsService(mockRepo)
		_, err := svc.GetAdminMetrics(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
