package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuditLogRepository struct {
	mock.Mock
}

func (m *MockAuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogRepository) FindAll(ctx context.Context, query dtos.AuditLogQuery) ([]models.AuditLog, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.AuditLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockAuditLogRepository) FindByID(ctx context.Context, id string) (*models.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestAuditLogService(t *testing.T) {
	ctx := context.Background()

	t.Run("Record success", func(t *testing.T) {
		repo := new(MockAuditLogRepository)
		repo.On("Create", ctx, mock.AnythingOfType("*models.AuditLog")).Return(nil)

		svc := NewAuditLogService(repo)
		req := dtos.CreateAuditLogRequest{
			ActorName:      "Budi Pratama",
			ActorRole:      "Senior Underwriter",
			Action:         "APPROVE_APPLICATION",
			Category:       "underwriting",
			TargetResource: "#APP-2026-8819",
			Details:        map[string]any{"note": "approved automatically"},
		}

		res, err := svc.Record(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, res.ID)
		assert.Equal(t, "Budi Pratama", res.ActorName)
		assert.Equal(t, "underwriting", res.Category)
		assert.Equal(t, "SUCCESS", res.Status)
		assert.NotEmpty(t, res.Hash)
		assert.Equal(t, "approved automatically", res.Details["note"])
		repo.AssertExpectations(t)
	})

	t.Run("Record validation failure", func(t *testing.T) {
		repo := new(MockAuditLogRepository)
		svc := NewAuditLogService(repo)
		req := dtos.CreateAuditLogRequest{
			ActorName: "",
		}

		res, err := svc.Record(ctx, req)
		require.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("List success", func(t *testing.T) {
		repo := new(MockAuditLogRepository)
		sampleLogs := []models.AuditLog{
			{
				ID:             "aud-1",
				Timestamp:      time.Now().UTC(),
				ActorName:      "Dewi Sartika",
				ActorRole:      "Product Actuary",
				Action:         "UPDATE_PRODUCT_PRICING",
				Category:       models.AuditCategoryProduct,
				TargetResource: "prod-1",
				IPAddress:      "127.0.0.1",
				Status:         models.AuditSeveritySuccess,
				Details:        `{"diff":"base_rate changed"}`,
				Hash:           "hash-xyz",
			},
		}
		repo.On("FindAll", ctx, mock.AnythingOfType("dtos.AuditLogQuery")).Return(sampleLogs, int64(1), nil)

		svc := NewAuditLogService(repo)
		listRes, err := svc.List(ctx, dtos.AuditLogQuery{Limit: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(1), listRes.Total)
		assert.Len(t, listRes.Data, 1)
		assert.Equal(t, "aud-1", listRes.Data[0].ID)
		assert.Equal(t, "base_rate changed", listRes.Data[0].Details["diff"])
		repo.AssertExpectations(t)
	})

	t.Run("GetByID success", func(t *testing.T) {
		repo := new(MockAuditLogRepository)
		sample := &models.AuditLog{
			ID:             "aud-1",
			Timestamp:      time.Now().UTC(),
			ActorName:      "Budi",
			ActorRole:      "Underwriter",
			Action:         "APPROVE",
			Category:       models.AuditCategoryUnderwriting,
			TargetResource: "#APP-1",
			IPAddress:      "127.0.0.1",
			Status:         models.AuditSeveritySuccess,
			Details:        "{}",
			Hash:           "hash-abc",
		}
		repo.On("FindByID", ctx, "aud-1").Return(sample, nil)

		svc := NewAuditLogService(repo)
		res, err := svc.GetByID(ctx, "aud-1")
		require.NoError(t, err)
		assert.Equal(t, "aud-1", res.ID)
		repo.AssertExpectations(t)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		repo := new(MockAuditLogRepository)
		repo.On("FindByID", ctx, "unknown").Return(nil, errors.New("audit log not found"))

		svc := NewAuditLogService(repo)
		res, err := svc.GetByID(ctx, "unknown")
		require.Error(t, err)
		assert.Nil(t, res)
		repo.AssertExpectations(t)
	})
}
