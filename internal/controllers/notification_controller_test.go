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

type fakeNotificationService struct {
	listResp       *dtos.NotificationListResponse
	listErr        error
	createResp     *dtos.NotificationResponse
	createErr      error
	markReadResp   *dtos.MarkReadResponse
	markReadErr    error
	markAllResp    *dtos.MarkAllReadResponse
	markAllErr     error
	unreadCountVal int64
	unreadCountErr error
}

func (f *fakeNotificationService) List(_ context.Context, _ dtos.NotificationQuery) (*dtos.NotificationListResponse, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResp, nil
}

func (f *fakeNotificationService) Create(_ context.Context, _ dtos.CreateNotificationRequest) (*dtos.NotificationResponse, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	return f.createResp, nil
}

func (f *fakeNotificationService) MarkAsRead(_ context.Context, id string) (*dtos.MarkReadResponse, error) {
	if f.markReadErr != nil {
		return nil, f.markReadErr
	}
	if id == "not-found" {
		return nil, constants.ErrNotificationNotFoundError
	}
	return f.markReadResp, nil
}

func (f *fakeNotificationService) MarkAllAsRead(_ context.Context) (*dtos.MarkAllReadResponse, error) {
	if f.markAllErr != nil {
		return nil, f.markAllErr
	}
	return f.markAllResp, nil
}

func (f *fakeNotificationService) GetUnreadCount(_ context.Context) (int64, error) {
	if f.unreadCountErr != nil {
		return 0, f.unreadCountErr
	}
	return f.unreadCountVal, nil
}

func setupNotificationApp(service *fakeNotificationService) *fiber.App {
	app := fiber.New()
	ctrl := controllers.NewNotificationController(service)

	app.Get("/api/v1/admin/notifications", ctrl.List)
	app.Post("/api/v1/admin/notifications", ctrl.Create)
	app.Patch("/api/v1/admin/notifications/:id/read", ctrl.MarkAsRead)
	app.Post("/api/v1/admin/notifications/mark-all-read", ctrl.MarkAllAsRead)
	app.Get("/api/v1/admin/notifications/unread-count", ctrl.GetUnreadCount)

	return app
}

func TestNotificationController_List(t *testing.T) {
	t.Run("returns 200 with list data", func(t *testing.T) {
		svc := &fakeNotificationService{
			listResp: &dtos.NotificationListResponse{
				Data: []dtos.NotificationResponse{
					{ID: "notif-1", Title: "Sample Alert", Severity: "INFO", IsRead: false},
				},
				Total:       1,
				UnreadCount: 1,
			},
		}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/notifications?category=underwriting", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, decodeErr)
		assert.Equal(t, float64(1), body["total"])
		assert.Equal(t, float64(1), body["unread_count"])
	})

	t.Run("returns 400 on validation failure", func(t *testing.T) {
		svc := &fakeNotificationService{
			listErr: errors.New(constants.ErrNotificationLimitInvalid),
		}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/notifications?limit=200", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 500 on internal service failure", func(t *testing.T) {
		svc := &fakeNotificationService{
			listErr: errors.New("db connection failure"),
		}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/notifications", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestNotificationController_Create(t *testing.T) {
	t.Run("returns 201 on valid request", func(t *testing.T) {
		svc := &fakeNotificationService{
			createResp: &dtos.NotificationResponse{
				ID:       "notif-new",
				Title:    "New Title",
				Type:     "APPLICATION_SUBMITTED",
				Severity: "INFO",
			},
		}
		app := setupNotificationApp(svc)

		payload := dtos.CreateNotificationRequest{
			Type:     "APPLICATION_SUBMITTED",
			Category: "underwriting",
			Severity: "INFO",
			Title:    "New Title",
			Message:  "Body content",
		}
		bytesData, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications", bytes.NewReader(bytesData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("returns 400 on invalid JSON body", func(t *testing.T) {
		svc := &fakeNotificationService{}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("returns 400 on validation error", func(t *testing.T) {
		svc := &fakeNotificationService{
			createErr: errors.New(constants.ErrNotificationTitleRequired),
		}
		app := setupNotificationApp(svc)

		payload := dtos.CreateNotificationRequest{Type: "SLA_WARNING"}
		bytesData, err := json.Marshal(payload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications", bytes.NewReader(bytesData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestNotificationController_MarkAsRead(t *testing.T) {
	t.Run("returns 200 on success", func(t *testing.T) {
		svc := &fakeNotificationService{
			markReadResp: &dtos.MarkReadResponse{
				ID:     "notif-1",
				IsRead: true,
				ReadAt: "2026-09-08T01:00:00Z",
			},
		}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/notifications/notif-1/read", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("returns 404 when item not found", func(t *testing.T) {
		svc := &fakeNotificationService{}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/notifications/not-found/read", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestNotificationController_MarkAllAsRead(t *testing.T) {
	t.Run("returns 200 on success", func(t *testing.T) {
		svc := &fakeNotificationService{
			markAllResp: &dtos.MarkAllReadResponse{
				UpdatedCount: 3,
				UnreadCount:  0,
			},
		}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/notifications/mark-all-read", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestNotificationController_GetUnreadCount(t *testing.T) {
	t.Run("returns 200 with unread count", func(t *testing.T) {
		svc := &fakeNotificationService{
			unreadCountVal: 3,
		}
		app := setupNotificationApp(svc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/notifications/unread-count", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, decodeErr)
		assert.Equal(t, float64(3), body["unread_count"])
	})
}
