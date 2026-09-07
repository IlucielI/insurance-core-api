package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/controllers"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSystemHealthService struct {
	overviewResp *dtos.SystemHealthOverviewResponse
	overviewErr  error
	servicesResp []dtos.ServiceHealthItem
	servicesErr  error
	routesResp   []dtos.RouteLatencyProbeItem
	routesErr    error
}

func (f *fakeSystemHealthService) GetOverview(_ context.Context) (*dtos.SystemHealthOverviewResponse, error) {
	if f.overviewErr != nil {
		return nil, f.overviewErr
	}
	return f.overviewResp, nil
}

func (f *fakeSystemHealthService) PingServices(_ context.Context, serviceID string) ([]dtos.ServiceHealthItem, error) {
	if f.servicesErr != nil {
		return nil, f.servicesErr
	}
	if serviceID != "" {
		res := make([]dtos.ServiceHealthItem, 0)
		for _, s := range f.servicesResp {
			if s.ID == serviceID {
				res = append(res, s)
			}
		}
		return res, nil
	}
	return f.servicesResp, nil
}

func (f *fakeSystemHealthService) PingRoutes(_ context.Context) ([]dtos.RouteLatencyProbeItem, error) {
	if f.routesErr != nil {
		return nil, f.routesErr
	}
	return f.routesResp, nil
}

func setupHealthApp(service *fakeSystemHealthService) *fiber.App {
	app := fiber.New()
	ctrl := controllers.NewAdminHealthController(service)

	app.Get("/api/v1/admin/health/overview", ctrl.GetOverview)
	app.Post("/api/v1/admin/health/ping", ctrl.PingServices)
	app.Post("/api/v1/admin/health/routes/ping", ctrl.PingRoutes)

	return app
}

func TestAdminHealthController_GetOverview(t *testing.T) {
	service := &fakeSystemHealthService{
		overviewResp: &dtos.SystemHealthOverviewResponse{
			OverallStatus:       dtos.ServiceHealthOnline,
			ActiveServicesCount: 5,
			TotalServicesCount:  5,
			AvgLatencyMs:        18.5,
			Services: []dtos.ServiceHealthItem{
				{ID: "service_core_api", Name: "Core API", Status: dtos.ServiceHealthOnline},
			},
			DatabaseStats: dtos.DatabasePoolStats{
				OpenConnections:    12,
				MaxOpenConnections: 50,
			},
		},
	}
	app := setupHealthApp(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health/overview", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	errDecode := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, errDecode)
	data, ok := body["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "online", data["overall_status"])

	t.Run("internal server error", func(t *testing.T) {
		errService := &fakeSystemHealthService{
			overviewErr: errors.New("database connection failed"),
		}
		errApp := setupHealthApp(errService)

		errReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/health/overview", nil)
		errResp, errTest := errApp.Test(errReq)
		require.NoError(t, errTest)
		assert.Equal(t, http.StatusInternalServerError, errResp.StatusCode)
	})
}

func TestAdminHealthController_PingServices(t *testing.T) {
	service := &fakeSystemHealthService{
		servicesResp: []dtos.ServiceHealthItem{
			{ID: "service_postgres", Name: "PostgreSQL", Status: dtos.ServiceHealthOnline, LatencyMs: 2.1},
			{ID: "service_redis", Name: "Redis", Status: dtos.ServiceHealthOnline, LatencyMs: 1.5},
		},
	}
	app := setupHealthApp(service)

	t.Run("ping all services", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/health/ping", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		errDecode := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, errDecode)
		data, ok := body["data"].([]any)
		require.True(t, ok)
		assert.Len(t, data, 2)
	})

	t.Run("ping single service", func(t *testing.T) {
		reqPayload, errMarshal := json.Marshal(dtos.PingServicesRequest{ServiceID: "service_postgres"})
		require.NoError(t, errMarshal)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/health/ping", bytes.NewReader(reqPayload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		errDecode := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, errDecode)
		data, ok := body["data"].([]any)
		require.True(t, ok)
		assert.Len(t, data, 1)
	})

	t.Run("ping error handling", func(t *testing.T) {
		errService := &fakeSystemHealthService{
			servicesErr: errors.New("ping timeout"),
		}
		errApp := setupHealthApp(errService)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/health/ping", nil)
		resp, err := errApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAdminHealthController_PingRoutes(t *testing.T) {
	service := &fakeSystemHealthService{
		routesResp: []dtos.RouteLatencyProbeItem{
			{ID: "route_health", Method: "GET", Path: "/health", LatencyMs: 1.2, StatusCode: 200},
		},
	}
	app := setupHealthApp(service)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/health/routes/ping", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	errDecode := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, errDecode)
	data, ok := body["data"].([]any)
	require.True(t, ok)
	assert.Len(t, data, 1)

	t.Run("route ping error handling", func(t *testing.T) {
		errService := &fakeSystemHealthService{
			routesErr: errors.New("probe failure"),
		}
		errApp := setupHealthApp(errService)

		errReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/health/routes/ping", nil)
		errResp, errTest := errApp.Test(errReq)
		require.NoError(t, errTest)
		assert.Equal(t, http.StatusInternalServerError, errResp.StatusCode)
	})
}
