package workers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) List(ctx context.Context, query dtos.NotificationQuery) (*dtos.NotificationListResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.NotificationListResponse), args.Error(1)
}

func (m *mockNotificationService) Create(ctx context.Context, req dtos.CreateNotificationRequest) (*dtos.NotificationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.NotificationResponse), args.Error(1)
}

func (m *mockNotificationService) CreateBatch(ctx context.Context, reqs []dtos.CreateNotificationRequest) ([]*dtos.NotificationResponse, error) {
	args := m.Called(ctx, reqs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dtos.NotificationResponse), args.Error(1)
}

func (m *mockNotificationService) MarkAsRead(ctx context.Context, id string) (*dtos.MarkReadResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.MarkReadResponse), args.Error(1)
}

func (m *mockNotificationService) MarkAllAsRead(ctx context.Context) (*dtos.MarkAllReadResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.MarkAllReadResponse), args.Error(1)
}

func (m *mockNotificationService) GetUnreadCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

type fakeSLASource struct {
	apps []models.Application
	err  error
}

func (f *fakeSLASource) FindPendingSLABreach(ctx context.Context, olderThan time.Time) ([]models.Application, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.apps, nil
}

func TestAdminNotificationWorkerLifecycle(t *testing.T) {
	sub := newFakeSubscriber()
	notifSvc := new(mockNotificationService)

	worker := NewAdminNotificationWorker(sub, notifSvc)
	err := worker.Start(context.Background())
	require.NoError(t, err)

	expectedTopics := []string{
		dtos.TopicApplicationSubmitted,
		dtos.TopicApplicationApproved,
		dtos.TopicApplicationRejected,
		dtos.TopicApplicationRFIRequested,
	}

	for _, topic := range expectedTopics {
		assert.Contains(t, sub.handlers, topic)
	}

	worker.Stop()
}

func TestAdminNotificationWorkerEventHandlers(t *testing.T) {
	ctx := context.Background()

	t.Run("handleApplicationSubmitted creates notification", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		notifSvc.On("Create", mock.Anything, mock.MatchedBy(func(req dtos.CreateNotificationRequest) bool {
			return req.Type == "APPLICATION_SUBMITTED" &&
				req.Category == "underwriting" &&
				req.Severity == "INFO" &&
				req.Link == "/queue" &&
				req.Title == "Aplikasi Baru: Budi Santoso (#APP-123)"
		})).Return(&dtos.NotificationResponse{ID: "notif-1"}, nil)

		worker := NewAdminNotificationWorker(sub, notifSvc)
		require.NoError(t, worker.Start(ctx))
		defer worker.Stop()

		event := dtos.ApplicationSubmittedEvent{
			ApplicationID: "APP-123",
			ProductName:   "Secure Life Plus",
			FullName:      "Budi Santoso",
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)

		handler := sub.handlers[dtos.TopicApplicationSubmitted]
		require.NotNil(t, handler)
		require.NoError(t, handler(ctx, data))

		notifSvc.AssertExpectations(t)
	})

	t.Run("handleApplicationApproved creates notification", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		notifSvc.On("Create", mock.Anything, mock.MatchedBy(func(req dtos.CreateNotificationRequest) bool {
			return req.Type == "APPLICATION_APPROVED" &&
				req.Category == "underwriting" &&
				req.Severity == "SUCCESS" &&
				req.Title == "Aplikasi Disetujui: #APP-123"
		})).Return(&dtos.NotificationResponse{ID: "notif-2"}, nil)

		worker := NewAdminNotificationWorker(sub, notifSvc)
		require.NoError(t, worker.Start(ctx))
		defer worker.Stop()

		event := dtos.ApplicationApprovedEvent{
			ApplicationID: "APP-123",
			ProductName:   "Secure Life Plus",
			FullName:      "Budi Santoso",
			ReviewedBy:    "Lead Underwriter",
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)

		handler := sub.handlers[dtos.TopicApplicationApproved]
		require.NotNil(t, handler)
		require.NoError(t, handler(ctx, data))

		notifSvc.AssertExpectations(t)
	})

	t.Run("handleApplicationRejected creates notification", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		notifSvc.On("Create", mock.Anything, mock.MatchedBy(func(req dtos.CreateNotificationRequest) bool {
			return req.Type == "APPLICATION_REJECTED" &&
				req.Category == "underwriting" &&
				req.Severity == "WARNING" &&
				req.Title == "Aplikasi Ditolak: #APP-123"
		})).Return(&dtos.NotificationResponse{ID: "notif-3"}, nil)

		worker := NewAdminNotificationWorker(sub, notifSvc)
		require.NoError(t, worker.Start(ctx))
		defer worker.Stop()

		event := dtos.ApplicationRejectedEvent{
			ApplicationID:   "APP-123",
			ProductName:     "Secure Life Plus",
			RejectionReason: "Medical check failed",
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)

		handler := sub.handlers[dtos.TopicApplicationRejected]
		require.NotNil(t, handler)
		require.NoError(t, handler(ctx, data))

		notifSvc.AssertExpectations(t)
	})

	t.Run("handleApplicationRFIRequested creates notification", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		notifSvc.On("Create", mock.Anything, mock.MatchedBy(func(req dtos.CreateNotificationRequest) bool {
			return req.Type == "APPLICATION_RFI" &&
				req.Category == "underwriting" &&
				req.Severity == "INFO" &&
				req.Title == "Permintaan Dokumen: #APP-123"
		})).Return(&dtos.NotificationResponse{ID: "notif-4"}, nil)

		worker := NewAdminNotificationWorker(sub, notifSvc)
		require.NoError(t, worker.Start(ctx))
		defer worker.Stop()

		event := dtos.ApplicationRFIRequestedEvent{
			ApplicationID: "APP-123",
			FullName:      "Budi Santoso",
			RequiredDocs:  []string{"KTP", "Slip Gaji"},
		}
		data, err := json.Marshal(event)
		require.NoError(t, err)

		handler := sub.handlers[dtos.TopicApplicationRFIRequested]
		require.NotNil(t, handler)
		require.NoError(t, handler(ctx, data))

		notifSvc.AssertExpectations(t)
	})
}

func TestAdminNotificationWorkerSLA(t *testing.T) {
	ctx := context.Background()

	t.Run("creates SLA warning when not yet alerted", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		slaSource := &fakeSLASource{
			apps: []models.Application{
				{
					ID:        "APP-BREACH-1",
					FullName:  "Dewi Sartika",
					CreatedAt: time.Now().Add(-25 * time.Hour),
				},
			},
		}

		unreadOnly := true
		notifSvc.On("List", ctx, dtos.NotificationQuery{
			UnreadOnly: &unreadOnly,
			Limit:      100,
		}).Return(&dtos.NotificationListResponse{
			Data: []dtos.NotificationResponse{},
		}, nil)

		notifSvc.On("CreateBatch", ctx, mock.MatchedBy(func(reqs []dtos.CreateNotificationRequest) bool {
			return len(reqs) == 1 &&
				reqs[0].Type == "SLA_WARNING" &&
				reqs[0].Category == "underwriting" &&
				reqs[0].Severity == "WARNING" &&
				reqs[0].Title == "SLA Warning: Aplikasi #APP-BREACH-1"
		})).Return([]*dtos.NotificationResponse{{ID: "notif-sla-1"}}, nil)

		worker := NewAdminNotificationWorker(sub, notifSvc, slaSource)
		worker.SetSLAThresholdHours(20)

		count, err := worker.CheckSLANow(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		notifSvc.AssertExpectations(t)
	})

	t.Run("skips creating SLA warning if unread alert already exists", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		slaSource := &fakeSLASource{
			apps: []models.Application{
				{
					ID:        "APP-BREACH-1",
					FullName:  "Dewi Sartika",
					CreatedAt: time.Now().Add(-25 * time.Hour),
				},
			},
		}

		unreadOnly := true
		notifSvc.On("List", ctx, dtos.NotificationQuery{
			UnreadOnly: &unreadOnly,
			Limit:      100,
		}).Return(&dtos.NotificationListResponse{
			Data: []dtos.NotificationResponse{
				{
					ID:     "notif-sla-existing",
					Title:  "SLA Warning: Aplikasi #APP-BREACH-1",
					IsRead: false,
				},
			},
		}, nil)

		worker := NewAdminNotificationWorker(sub, notifSvc, slaSource)
		worker.SetSLAThresholdHours(20)

		count, err := worker.CheckSLANow(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		notifSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		notifSvc.AssertNotCalled(t, "CreateBatch", mock.Anything, mock.Anything)
		notifSvc.AssertExpectations(t)
	})

	t.Run("guard test against N+1: processes multiple breach items in a single batch query", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		// 5 pending applications
		apps := []models.Application{
			{ID: "APP-1", FullName: "User 1", CreatedAt: time.Now().Add(-21 * time.Hour)},
			{ID: "APP-2", FullName: "User 2", CreatedAt: time.Now().Add(-22 * time.Hour)},
			{ID: "APP-3", FullName: "User 3", CreatedAt: time.Now().Add(-23 * time.Hour)},
			{ID: "APP-4", FullName: "User 4", CreatedAt: time.Now().Add(-24 * time.Hour)},
			{ID: "APP-5", FullName: "User 5", CreatedAt: time.Now().Add(-25 * time.Hour)},
		}
		slaSource := &fakeSLASource{apps: apps}

		unreadOnly := true
		// APP-1 already alerted, others are not
		notifSvc.On("List", ctx, dtos.NotificationQuery{
			UnreadOnly: &unreadOnly,
			Limit:      100,
		}).Return(&dtos.NotificationListResponse{
			Data: []dtos.NotificationResponse{
				{ID: "notif-app-1", Title: "SLA Warning: Aplikasi #APP-1", IsRead: false},
			},
		}, nil).Once()

		// Exactly ONE batch call for all 4 new applications, NOT 4 separate calls
		notifSvc.On("CreateBatch", ctx, mock.MatchedBy(func(reqs []dtos.CreateNotificationRequest) bool {
			return len(reqs) == 4
		})).Return([]*dtos.NotificationResponse{
			{ID: "n2"}, {ID: "n3"}, {ID: "n4"}, {ID: "n5"},
		}, nil).Once()

		worker := NewAdminNotificationWorker(sub, notifSvc, slaSource)
		worker.SetSLAThresholdHours(20)

		count, err := worker.CheckSLANow(ctx)
		require.NoError(t, err)
		assert.Equal(t, 4, count)

		// Assert that no per-item Create is called
		notifSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		// Assert that CreateBatch is called exactly ONCE
		notifSvc.AssertNumberOfCalls(t, "CreateBatch", 1)
		notifSvc.AssertNumberOfCalls(t, "List", 1)
		notifSvc.AssertExpectations(t)
	})

	t.Run("handles empty breach collection without any query or batch calls", func(t *testing.T) {
		sub := newFakeSubscriber()
		notifSvc := new(mockNotificationService)

		slaSource := &fakeSLASource{apps: []models.Application{}}
		worker := NewAdminNotificationWorker(sub, notifSvc, slaSource)

		count, err := worker.CheckSLANow(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		notifSvc.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
		notifSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		notifSvc.AssertNotCalled(t, "CreateBatch", mock.Anything, mock.Anything)
	})
}
