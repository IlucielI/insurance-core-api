package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	FindAll(ctx context.Context, query dtos.NotificationQuery) ([]models.Notification, int64, int64, error)
	FindByID(ctx context.Context, id string) (*models.Notification, error)
	MarkAsRead(ctx context.Context, id string, readAt time.Time) (*models.Notification, error)
	MarkAllAsRead(ctx context.Context, readAt time.Time) (int64, error)
	CountUnread(ctx context.Context) (int64, error)
}

type PostgresNotificationRepository struct {
	db *gorm.DB
}

func NewPostgresNotificationRepository(db *gorm.DB) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *PostgresNotificationRepository) FindAll(ctx context.Context, query dtos.NotificationQuery) ([]models.Notification, int64, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.Notification{})

	if query.Category != "" && query.Category != "all" {
		db = db.Where("category = ?", query.Category)
	}

	if query.Severity != "" && query.Severity != "all" && query.Severity != "ALL" {
		db = db.Where("severity = ?", query.Severity)
	}

	if query.UnreadOnly != nil && *query.UnreadOnly {
		db = db.Where("is_read = ?", false)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	unreadCount, err := r.CountUnread(ctx)
	if err != nil {
		return nil, 0, 0, err
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var items []models.Notification
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, 0, err
	}

	return items, total, unreadCount, nil
}

func (r *PostgresNotificationRepository) FindByID(ctx context.Context, id string) (*models.Notification, error) {
	var item models.Notification
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrNotificationNotFoundError
		}
		return nil, err
	}
	return &item, nil
}

func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, id string, readAt time.Time) (*models.Notification, error) {
	res := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"is_read": true,
			"read_at": readAt,
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, constants.ErrNotificationNotFoundError
	}

	return r.FindByID(ctx, id)
}

func (r *PostgresNotificationRepository) MarkAllAsRead(ctx context.Context, readAt time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&models.Notification{}).
		Where("is_read = ?", false).
		Updates(map[string]any{
			"is_read": true,
			"read_at": readAt,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (r *PostgresNotificationRepository) CountUnread(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Notification{}).Where("is_read = ?", false).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
