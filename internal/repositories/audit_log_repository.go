package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *models.AuditLog) error
	FindAll(ctx context.Context, query dtos.AuditLogQuery) ([]models.AuditLog, int64, error)
	FindByID(ctx context.Context, id string) (*models.AuditLog, error)
	Count(ctx context.Context) (int64, error)
}

type PostgresAuditLogRepository struct {
	db *gorm.DB
}

func NewPostgresAuditLogRepository(db *gorm.DB) *PostgresAuditLogRepository {
	return &PostgresAuditLogRepository{db: db}
}

func (r *PostgresAuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *PostgresAuditLogRepository) FindAll(ctx context.Context, query dtos.AuditLogQuery) ([]models.AuditLog, int64, error) {
	db := r.db.WithContext(ctx).Model(&models.AuditLog{})

	if query.Category != "" && query.Category != "all" {
		db = db.Where("category = ?", query.Category)
	}

	if query.Status != "" && query.Status != "all" && query.Status != "ALL" {
		db = db.Where("status = ?", query.Status)
	}

	if query.Search != "" {
		pattern := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where(
			"LOWER(actor_name) LIKE ? OR LOWER(target_resource) LIKE ? OR LOWER(action) LIKE ? OR ip_address LIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}

	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	var logs []models.AuditLog
	if err := db.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *PostgresAuditLogRepository) FindByID(ctx context.Context, id string) (*models.AuditLog, error) {
	var log models.AuditLog
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrAuditLogNotFoundError
		}
		return nil, err
	}
	return &log, nil
}

func (r *PostgresAuditLogRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
