package services

import (
	"context"
	"math"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type MetricsService interface {
	GetAdminMetrics(ctx context.Context) (dtos.AdminMetricsResponse, error)
}

type metricsService struct {
	repo repositories.MetricsRepository
}

func NewMetricsService(repo repositories.MetricsRepository) MetricsService {
	return &metricsService{repo: repo}
}

func (s *metricsService) GetAdminMetrics(ctx context.Context) (dtos.AdminMetricsResponse, error) {
	raw, err := s.repo.GetAdminMetrics(ctx)
	if err != nil {
		return dtos.AdminMetricsResponse{}, err
	}

	var appMetrics dtos.ApplicationMetricsDTO
	for _, sc := range raw.StatusCounts {
		appMetrics.Total += sc.Count
		switch sc.Status {
		case models.ApplicationStatusDraft:
			appMetrics.Draft = sc.Count
		case models.ApplicationStatusSubmitted:
			appMetrics.Submitted = sc.Count
		case models.ApplicationStatusUnderReview:
			appMetrics.UnderReview = sc.Count
		case models.ApplicationStatusApproved:
			appMetrics.Approved = sc.Count
		case models.ApplicationStatusRejected:
			appMetrics.Rejected = sc.Count
		}
	}

	totalDecided := appMetrics.Approved + appMetrics.Rejected
	if totalDecided > 0 {
		appMetrics.ApprovalRate = math.Round((float64(appMetrics.Approved)/float64(totalDecided))*1000) / 10
	} else if appMetrics.Total > 0 {
		appMetrics.ApprovalRate = math.Round((float64(appMetrics.Approved)/float64(appMetrics.Total))*1000) / 10
	}

	premiumMetrics := dtos.PremiumMetricsDTO{
		TotalWrittenPremium: raw.TotalWrittenPremium,
		ActivePoliciesCount: raw.ActivePoliciesCount,
	}

	var avgSLA float64
	var complianceRate float64
	if raw.SLAMetrics.TotalReviewed > 0 {
		avgSLA = math.Round((raw.SLAMetrics.TotalDurationMins/float64(raw.SLAMetrics.TotalReviewed))*10) / 10
		complianceRate = math.Round((float64(raw.SLAMetrics.TotalWithinSLA)/float64(raw.SLAMetrics.TotalReviewed))*1000) / 10
	} else {
		complianceRate = 100.0
	}

	underwritingMetrics := dtos.UnderwritingMetricsDTO{
		PendingReviewCount: appMetrics.UnderReview + appMetrics.Submitted,
		AverageSLAMinutes:  avgSLA,
		SLATargetMinutes:   60,
		SLAComplianceRate:  complianceRate,
	}

	topProducts := make([]dtos.ProductPerformanceMetricDTO, 0, len(raw.TopProducts))
	for _, tp := range raw.TopProducts {
		topProducts = append(topProducts, dtos.ProductPerformanceMetricDTO{
			ProductID:           tp.ProductID,
			ProductSlug:         tp.ProductSlug,
			ProductName:         tp.ProductName,
			Category:            string(tp.Category),
			ActivePoliciesCount: tp.ActivePoliciesCount,
			TotalPremium:        tp.TotalPremium,
			LossRatio:           0.15,
		})
	}

	return dtos.AdminMetricsResponse{
		Applications: appMetrics,
		Premiums:     premiumMetrics,
		Underwriting: underwritingMetrics,
		TopProducts:  topProducts,
	}, nil
}
