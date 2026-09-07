package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, item *models.Notification) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockNotificationRepository) CreateBatch(ctx context.Context, items []models.Notification) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

func (m *MockNotificationRepository) FindAll(ctx context.Context, query dtos.NotificationQuery) ([]models.Notification, int64, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Get(2).(int64), args.Error(3)
	}
	return args.Get(0).([]models.Notification), args.Get(1).(int64), args.Get(2).(int64), args.Error(3)
}

func (m *MockNotificationRepository) FindByID(ctx context.Context, id string) (*models.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, id string, readAt time.Time) (*models.Notification, error) {
	args := m.Called(ctx, id, readAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, readAt time.Time) (int64, error) {
	args := m.Called(ctx, readAt)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) CountUnread(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestNotificationService(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		repo.On("Create", ctx, mock.AnythingOfType("*models.Notification")).Return(nil)

		svc := NewNotificationService(repo)
		req := dtos.CreateNotificationRequest{
			Type:     "APPLICATION_SUBMITTED",
			Category: "underwriting",
			Severity: "INFO",
			Title:    "New Application Received",
			Message:  "Application #APP-1 waiting for review",
			Link:     "/underwriting",
		}

		resp, err := svc.Create(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, req.Title, resp.Title)
		assert.Equal(t, req.Type, resp.Type)
		assert.Equal(t, req.Category, resp.Category)
		assert.Equal(t, "INFO", resp.Severity)
		assert.False(t, resp.IsRead)
		repo.AssertExpectations(t)
	})

	t.Run("Create validation error", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		svc := NewNotificationService(repo)

		req := dtos.CreateNotificationRequest{
			Type: "APPLICATION_SUBMITTED",
		}

		resp, err := svc.Create(ctx, req)
		assert.EqualError(t, err, constants.ErrNotificationTitleRequired)
		assert.Nil(t, resp)
	})

	t.Run("List success", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		now := time.Now().UTC()
		items := []models.Notification{
			{
				ID:        "notif-1",
				Type:      "APPLICATION_SUBMITTED",
				Category:  models.NotificationCategoryUnderwriting,
				Severity:  models.NotificationSeverityInfo,
				Title:     "Title 1",
				Message:   "Message 1",
				CreatedAt: now,
			},
		}
		repo.On("FindAll", ctx, mock.AnythingOfType("dtos.NotificationQuery")).Return(items, int64(1), int64(1), nil)

		svc := NewNotificationService(repo)
		query := dtos.NotificationQuery{Limit: 10}

		result, err := svc.List(ctx, query)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.Total)
		assert.Equal(t, int64(1), result.UnreadCount)
		assert.Len(t, result.Data, 1)
		assert.Equal(t, "notif-1", result.Data[0].ID)
		repo.AssertExpectations(t)
	})

	t.Run("List validation error", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		svc := NewNotificationService(repo)

		query := dtos.NotificationQuery{Limit: 1000}
		result, err := svc.List(ctx, query)
		assert.EqualError(t, err, constants.ErrNotificationLimitInvalid)
		assert.Nil(t, result)
	})

	t.Run("MarkAsRead success", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		readAt := time.Now().UTC()
		mockItem := &models.Notification{
			ID:     "notif-1",
			IsRead: true,
			ReadAt: &readAt,
		}
		repo.On("MarkAsRead", ctx, "notif-1", mock.AnythingOfType("time.Time")).Return(mockItem, nil)

		svc := NewNotificationService(repo)
		resp, err := svc.MarkAsRead(ctx, "notif-1")
		require.NoError(t, err)
		assert.Equal(t, "notif-1", resp.ID)
		assert.True(t, resp.IsRead)
		assert.NotEmpty(t, resp.ReadAt)
		repo.AssertExpectations(t)
	})

	t.Run("MarkAsRead failure", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		repo.On("MarkAsRead", ctx, "missing-id", mock.AnythingOfType("time.Time")).Return(nil, errors.New("not found"))

		svc := NewNotificationService(repo)
		resp, err := svc.MarkAsRead(ctx, "missing-id")
		assert.Error(t, err)
		assert.Nil(t, resp)
		repo.AssertExpectations(t)
	})

	t.Run("MarkAllAsRead success", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		repo.On("MarkAllAsRead", ctx, mock.AnythingOfType("time.Time")).Return(int64(4), nil)
		repo.On("CountUnread", ctx).Return(int64(0), nil)

		svc := NewNotificationService(repo)
		resp, err := svc.MarkAllAsRead(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(4), resp.UpdatedCount)
		assert.Equal(t, int64(0), resp.UnreadCount)
		repo.AssertExpectations(t)
	})

	t.Run("GetUnreadCount success", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		repo.On("CountUnread", ctx).Return(int64(5), nil)

		svc := NewNotificationService(repo)
		count, err := svc.GetUnreadCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), count)
		repo.AssertExpectations(t)
	})

	t.Run("CreateBatch success", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		reqs := []dtos.CreateNotificationRequest{
			{
				Type:     "SLA_WARNING",
				Category: "underwriting",
				Severity: "WARNING",
				Title:    "SLA Warning: App 1",
				Message:  "Pending",
			},
			{
				Type:     "SLA_WARNING",
				Category: "underwriting",
				Severity: "WARNING",
				Title:    "SLA Warning: App 2",
				Message:  "Pending",
			},
		}

		repo.On("CreateBatch", ctx, mock.MatchedBy(func(items []models.Notification) bool {
			return len(items) == 2 && items[0].Title == "SLA Warning: App 1" && items[1].Title == "SLA Warning: App 2"
		})).Return(nil)

		svc := NewNotificationService(repo)
		resps, err := svc.CreateBatch(ctx, reqs)
		require.NoError(t, err)
		assert.Len(t, resps, 2)
		assert.Equal(t, "SLA Warning: App 1", resps[0].Title)
		assert.Equal(t, "SLA Warning: App 2", resps[1].Title)
		repo.AssertExpectations(t)
	})

	t.Run("CreateBatch empty returns empty slice without calling repo", func(t *testing.T) {
		repo := new(MockNotificationRepository)
		svc := NewNotificationService(repo)
		resps, err := svc.CreateBatch(ctx, []dtos.CreateNotificationRequest{})
		require.NoError(t, err)
		assert.Empty(t, resps)
		repo.AssertNotCalled(t, "CreateBatch", mock.Anything, mock.Anything)
	})
}
