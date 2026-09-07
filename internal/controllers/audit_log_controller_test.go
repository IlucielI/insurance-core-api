package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/controllers"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuditLogService struct {
	listResp   *dtos.AuditLogListResponse
	listErr    error
	getResp    *dtos.AuditLogResponse
	getErr     error
	recordResp *dtos.AuditLogResponse
	recordErr  error
}

func (f *fakeAuditLogService) List(_ context.Context, _ dtos.AuditLogQuery) (*dtos.AuditLogListResponse, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResp, nil
}

func (f *fakeAuditLogService) GetByID(_ context.Context, id string) (*dtos.AuditLogResponse, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	if id == "not-found" {
		return nil, constants.ErrAuditLogNotFoundError
	}
	return f.getResp, nil
}

func (f *fakeAuditLogService) Record(_ context.Context, _ dtos.CreateAuditLogRequest) (*dtos.AuditLogResponse, error) {
	if f.recordErr != nil {
		return nil, f.recordErr
	}
	return f.recordResp, nil
}

func setupAuditApp(service *fakeAuditLogService) *fiber.App {
	app := fiber.New()
	ctrl := controllers.NewAuditLogController(service)

	app.Get("/api/v1/admin/audit-logs", ctrl.List)
	app.Get("/api/v1/admin/audit-logs/:id", ctrl.GetByID)
	app.Post("/api/v1/admin/audit-logs", ctrl.Create)

	return app
}

func TestAuditLogController_List(t *testing.T) {
	service := &fakeAuditLogService{
		listResp: &dtos.AuditLogListResponse{
			Data: []dtos.AuditLogResponse{
				{ID: "aud-1", ActorName: "Budi", Category: "underwriting", Status: "SUCCESS"},
			},
			Total: 1,
		},
	}
	app := setupAuditApp(service)

	t.Run("successful list query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?category=underwriting", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		errDecode := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, errDecode)
		data, ok := body["data"].([]any)
		require.True(t, ok)
		assert.Len(t, data, 1)
		assert.Equal(t, float64(1), body["total"])
	})

	t.Run("validation error on query", func(t *testing.T) {
		errService := &fakeAuditLogService{
			listErr: errors.New(constants.ErrAuditLogCategoryInvalid),
		}
		errApp := setupAuditApp(errService)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs?category=invalid", nil)
		resp, err := errApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("server error on list", func(t *testing.T) {
		errService := &fakeAuditLogService{
			listErr: errors.New("db failure"),
		}
		errApp := setupAuditApp(errService)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
		resp, err := errApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAuditLogController_GetByID(t *testing.T) {
	service := &fakeAuditLogService{
		getResp: &dtos.AuditLogResponse{
			ID:        "aud-1",
			ActorName: "Budi",
		},
	}
	app := setupAuditApp(service)

	t.Run("found by identifier", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/aud-1", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		errDecode := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, errDecode)
		data, ok := body["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "aud-1", data["id"])
	})

	t.Run("missing record returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/not-found", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("internal server error on retrieval", func(t *testing.T) {
		errService := &fakeAuditLogService{
			getErr: errors.New("internal query error"),
		}
		errApp := setupAuditApp(errService)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs/aud-err", nil)
		resp, err := errApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAuditLogController_Create(t *testing.T) {
	service := &fakeAuditLogService{
		recordResp: &dtos.AuditLogResponse{
			ID:        "aud-new",
			ActorName: "Budi",
			Status:    "SUCCESS",
		},
	}
	app := setupAuditApp(service)

	t.Run("successful audit creation", func(t *testing.T) {
		payload, errMarshal := json.Marshal(dtos.CreateAuditLogRequest{
			ActorName:      "Budi",
			ActorRole:      "Senior Underwriter",
			Action:         "APPROVE",
			Category:       "underwriting",
			TargetResource: "#APP-1",
		})
		require.NoError(t, errMarshal)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit-logs", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("invalid json body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit-logs", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("validation error on creation", func(t *testing.T) {
		errService := &fakeAuditLogService{
			recordErr: errors.New(constants.ErrAuditLogActorNameRequired),
		}
		errApp := setupAuditApp(errService)

		payload, errMarshal := json.Marshal(dtos.CreateAuditLogRequest{})
		require.NoError(t, errMarshal)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit-logs", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := errApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("server error on creation", func(t *testing.T) {
		errService := &fakeAuditLogService{
			recordErr: errors.New("db write failed"),
		}
		errApp := setupAuditApp(errService)

		payload, errMarshal := json.Marshal(dtos.CreateAuditLogRequest{ActorName: "Budi"})
		require.NoError(t, errMarshal)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/audit-logs", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		resp, err := errApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
