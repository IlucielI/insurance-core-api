package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresAuditLogRepository(t *testing.T) {
	db := sqliteDB(t)
	repo := NewPostgresAuditLogRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	log1 := models.AuditLog{
		ID:             "aud-1",
		Timestamp:      now,
		ActorName:      "Budi Pratama",
		ActorRole:      "Senior Underwriter",
		Action:         "APPROVE_APPLICATION",
		Category:       models.AuditCategoryUnderwriting,
		TargetResource: "#APP-2026-8819",
		IPAddress:      "192.168.10.45",
		Status:         models.AuditSeveritySuccess,
		Details:        `{"applicant":"Budi Santoso"}`,
		Hash:           "hash123",
	}

	log2 := models.AuditLog{
		ID:             "aud-2",
		Timestamp:      now.Add(-time.Hour),
		ActorName:      "Dewi Sartika",
		ActorRole:      "Product Actuary",
		Action:         "UPDATE_PRODUCT_PRICING",
		Category:       models.AuditCategoryProduct,
		TargetResource: "prod_secure_life_plus",
		IPAddress:      "192.168.10.12",
		Status:         models.AuditSeveritySuccess,
		Details:        `{"product":"Secure Life Plus"}`,
		Hash:           "hash456",
	}

	log3 := models.AuditLog{
		ID:             "aud-3",
		Timestamp:      now.Add(-2 * time.Hour),
		ActorName:      "Andi Wijaya",
		ActorRole:      "Junior Underwriter",
		Action:         "SECURITY_PIN_FAILURE",
		Category:       models.AuditCategoryAuth,
		TargetResource: "#APP-2026-8818",
		IPAddress:      "192.168.10.68",
		Status:         models.AuditSeverityFailed,
		Details:        `{"attempt":1}`,
		Hash:           "hash789",
	}

	require.NoError(t, repo.Create(ctx, &log1))
	require.NoError(t, repo.Create(ctx, &log2))
	require.NoError(t, repo.Create(ctx, &log3))

	t.Run("Count", func(t *testing.T) {
		count, err := repo.Count(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("FindByID success", func(t *testing.T) {
		found, err := repo.FindByID(ctx, "aud-1")
		require.NoError(t, err)
		assert.Equal(t, "Budi Pratama", found.ActorName)
		assert.Equal(t, "#APP-2026-8819", found.TargetResource)
	})

	t.Run("FindByID not found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, "non-existent")
		require.Error(t, err)
		assert.Nil(t, found)
		assert.Equal(t, constants.ErrAuditLogNotFoundError, err)
	})

	t.Run("FindAll with no filters", func(t *testing.T) {
		logs, total, err := repo.FindAll(ctx, dtos.AuditLogQuery{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, logs, 3)
		assert.Equal(t, "aud-1", logs[0].ID) // ordered by timestamp desc
	})

	t.Run("FindAll filter by category", func(t *testing.T) {
		logs, total, err := repo.FindAll(ctx, dtos.AuditLogQuery{Category: "underwriting"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
		assert.Equal(t, "aud-1", logs[0].ID)
	})

	t.Run("FindAll filter by status", func(t *testing.T) {
		logs, total, err := repo.FindAll(ctx, dtos.AuditLogQuery{Status: "FAILED"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
		assert.Equal(t, "aud-3", logs[0].ID)
	})

	t.Run("FindAll search query", func(t *testing.T) {
		logs, total, err := repo.FindAll(ctx, dtos.AuditLogQuery{Search: "Sartika"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
		assert.Equal(t, "aud-2", logs[0].ID)
	})

	t.Run("FindAll pagination", func(t *testing.T) {
		logs, total, err := repo.FindAll(ctx, dtos.AuditLogQuery{Limit: 1, Offset: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, logs, 1)
		assert.Equal(t, "aud-2", logs[0].ID)
	})
}
