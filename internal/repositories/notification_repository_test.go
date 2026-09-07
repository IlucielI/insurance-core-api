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

func TestPostgresNotificationRepository(t *testing.T) {
	db := sqliteDB(t)
	repo := NewPostgresNotificationRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	item1 := models.Notification{
		ID:        "notif-1",
		Type:      "APPLICATION_SUBMITTED",
		Category:  models.NotificationCategoryUnderwriting,
		Severity:  models.NotificationSeverityInfo,
		Title:     "New Application",
		Message:   "Application #APP-1 submitted",
		Link:      "/underwriting",
		IsRead:    false,
		CreatedAt: now,
	}

	item2 := models.Notification{
		ID:        "notif-2",
		Type:      "SLA_WARNING",
		Category:  models.NotificationCategoryUnderwriting,
		Severity:  models.NotificationSeverityWarning,
		Title:     "SLA Warning",
		Message:   "Application #APP-2 pending for 24h",
		Link:      "/underwriting",
		IsRead:    false,
		CreatedAt: now.Add(-time.Hour),
	}

	item3 := models.Notification{
		ID:        "notif-3",
		Type:      "SYSTEM_ALERT",
		Category:  models.NotificationCategorySystem,
		Severity:  models.NotificationSeveritySuccess,
		Title:     "System Health Check",
		Message:   "All services running normally",
		Link:      "/health",
		IsRead:    true,
		CreatedAt: now.Add(-2 * time.Hour),
	}

	require.NoError(t, repo.Create(ctx, &item1))
	require.NoError(t, repo.Create(ctx, &item2))
	require.NoError(t, repo.Create(ctx, &item3))

	t.Run("CountUnread", func(t *testing.T) {
		count, err := repo.CountUnread(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("FindByID found", func(t *testing.T) {
		item, err := repo.FindByID(ctx, "notif-1")
		require.NoError(t, err)
		assert.Equal(t, "notif-1", item.ID)
		assert.Equal(t, "New Application", item.Title)
		assert.False(t, item.IsRead)
	})

	t.Run("FindByID not found", func(t *testing.T) {
		item, err := repo.FindByID(ctx, "nonexistent")
		assert.ErrorIs(t, err, constants.ErrNotificationNotFoundError)
		assert.Nil(t, item)
	})

	t.Run("FindAll unfiltered", func(t *testing.T) {
		items, total, unread, err := repo.FindAll(ctx, dtos.NotificationQuery{})
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Equal(t, int64(2), unread)
		assert.Len(t, items, 3)
	})

	t.Run("FindAll unread only", func(t *testing.T) {
		unreadOnly := true
		items, total, unread, err := repo.FindAll(ctx, dtos.NotificationQuery{
			UnreadOnly: &unreadOnly,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, int64(2), unread)
		assert.Len(t, items, 2)
	})

	t.Run("FindAll filter by category", func(t *testing.T) {
		items, total, _, err := repo.FindAll(ctx, dtos.NotificationQuery{
			Category: "system",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, items, 1)
		assert.Equal(t, "notif-3", items[0].ID)
	})

	t.Run("FindAll filter by severity", func(t *testing.T) {
		items, total, _, err := repo.FindAll(ctx, dtos.NotificationQuery{
			Severity: "WARNING",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, items, 1)
		assert.Equal(t, "notif-2", items[0].ID)
	})

	t.Run("MarkAsRead item", func(t *testing.T) {
		readTime := time.Now().UTC()
		marked, err := repo.MarkAsRead(ctx, "notif-1", readTime)
		require.NoError(t, err)
		assert.True(t, marked.IsRead)
		assert.NotNil(t, marked.ReadAt)

		unreadAfter, err := repo.CountUnread(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), unreadAfter)
	})

	t.Run("MarkAsRead non-existent item", func(t *testing.T) {
		readTime := time.Now().UTC()
		marked, err := repo.MarkAsRead(ctx, "not-found-id", readTime)
		assert.ErrorIs(t, err, constants.ErrNotificationNotFoundError)
		assert.Nil(t, marked)
	})

	t.Run("MarkAllAsRead", func(t *testing.T) {
		readTime := time.Now().UTC()
		affected, err := repo.MarkAllAsRead(ctx, readTime)
		require.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		unreadAfter, err := repo.CountUnread(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(0), unreadAfter)
	})
}
