package services

import (
	"context"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) GetJSON(ctx context.Context, key string, dest any) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCache) DeletePrefix(ctx context.Context, prefix string) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, keys ...string) (bool, error) {
	args := m.Called(ctx, keys)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestSystemHealthService(t *testing.T) {
	ctx := context.Background()

	t.Run("GetOverview returns aggregated stats", func(t *testing.T) {
		auditRepo := new(MockAuditLogRepository)
		sampleLogs := []models.AuditLog{
			{
				ID:             "aud-1",
				Timestamp:      time.Now().UTC(),
				ActorName:      "Budi",
				ActorRole:      "Underwriter",
				Action:         "APPROVE",
				Category:       models.AuditCategoryUnderwriting,
				TargetResource: "#APP-1",
				IPAddress:      "127.0.0.1",
				Status:         models.AuditSeveritySuccess,
				Details:        `{"key":"val"}`,
				Hash:           "hash-abc",
			},
		}
		auditRepo.On("FindAll", ctx, dtos.AuditLogQuery{Limit: 10, Offset: 0}).Return(sampleLogs, int64(1), nil)

		cache := new(MockCache)
		cache.On("Ping", mock.Anything).Return(nil)

		svc := NewSystemHealthService(nil, cache, auditRepo, config.Config{
			HTTPPort:    "8080",
			Version:     "v1.2.0",
			GitHash:     "9a4f2b1",
			DatabaseURL: "postgres://insurance:insurance@172.17.0.1:5432/insurance_core?sslmode=disable",
			RedisHost:   "172.17.0.1",
			RedisPort:   6379,
			SMTPHost:    "172.17.0.1",
			SMTPPort:    1025,
		}, time.Now().UTC())

		res, err := svc.GetOverview(ctx)
		require.NoError(t, err)
		assert.Equal(t, dtos.ServiceHealthOnline, res.OverallStatus)
		assert.Equal(t, 4, res.ActiveServicesCount)
		assert.Equal(t, 4, res.TotalServicesCount)
		assert.Greater(t, res.AvgLatencyMs, 0.0)
		assert.Len(t, res.RecentAuditLogs, 1)
		assert.Equal(t, 50, res.DatabaseStats.MaxOpenConnections)
	})

	t.Run("PingServices returns single service by ID", func(t *testing.T) {
		svc := NewSystemHealthService(nil, nil, nil, config.Config{
			HTTPPort:    "8080",
			DatabaseURL: "postgres://insurance:insurance@172.17.0.1:5432/insurance_core?sslmode=disable",
			RedisHost:   "172.17.0.1",
			RedisPort:   6379,
			SMTPHost:    "172.17.0.1",
			SMTPPort:    1025,
		}, time.Now().UTC())

		services, err := svc.PingServices(ctx, "service_postgres")
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "service_postgres", services[0].ID)
		assert.Equal(t, "PostgreSQL 16 & pgvector DB", services[0].Name)
		assert.Equal(t, "172.17.0.1:5432/insurance_core", services[0].Endpoint)
	})

	t.Run("PingRoutes returns route latency probes", func(t *testing.T) {
		svc := NewSystemHealthService(nil, nil, nil, config.Config{
			HTTPPort: "8080",
		}, time.Now().UTC())

		routes, err := svc.PingRoutes(ctx)
		require.NoError(t, err)
		assert.Len(t, routes, 6)
		assert.Equal(t, "route_health", routes[0].ID)
		assert.Equal(t, 200, routes[0].StatusCode)
	})
}
