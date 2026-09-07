package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/gofiber/fiber/v2"
)

type mockMetricsService struct {
	res dtos.AdminMetricsResponse
	err error
}

func (m *mockMetricsService) GetAdminMetrics(ctx context.Context) (dtos.AdminMetricsResponse, error) {
	return m.res, m.err
}

func TestAdminMetricsController_GetMetrics(t *testing.T) {
	t.Run("success returns 200 with data", func(t *testing.T) {
		svc := &mockMetricsService{
			res: dtos.AdminMetricsResponse{
				Applications: dtos.ApplicationMetricsDTO{
					Total:        1284,
					Approved:     1142,
					ApprovalRate: 94.2,
				},
				Premiums: dtos.PremiumMetricsDTO{
					TotalWrittenPremium: 8420000000,
					ActivePoliciesCount: 1142,
				},
				Underwriting: dtos.UnderwritingMetricsDTO{
					PendingReviewCount: 28,
					AverageSLAMinutes:  18.4,
					SLATargetMinutes:   60,
					SLAComplianceRate:  98.4,
				},
			},
		}

		app := fiber.New()
		ctrl := NewAdminMetricsController(svc)
		app.Get("/api/v1/admin/metrics", ctrl.GetMetrics)

		req := httptest.NewRequest("GET", "/api/v1/admin/metrics", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}
		var jsonRes struct {
			Data dtos.AdminMetricsResponse `json:"data"`
		}
		if err := json.Unmarshal(body, &jsonRes); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if jsonRes.Data.Applications.Total != 1284 {
			t.Errorf("Total = %d, want 1284", jsonRes.Data.Applications.Total)
		}
		if jsonRes.Data.Premiums.TotalWrittenPremium != 8420000000 {
			t.Errorf("TotalWrittenPremium = %d, want 8420000000", jsonRes.Data.Premiums.TotalWrittenPremium)
		}
	})

	t.Run("service error returns 500", func(t *testing.T) {
		svc := &mockMetricsService{
			err: errors.New("db error"),
		}

		app := fiber.New()
		ctrl := NewAdminMetricsController(svc)
		app.Get("/api/v1/admin/metrics", ctrl.GetMetrics)

		req := httptest.NewRequest("GET", "/api/v1/admin/metrics", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", resp.StatusCode)
		}
	})
}
