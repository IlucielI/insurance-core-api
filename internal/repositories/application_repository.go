package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

var ErrApplicationNotFound = errors.New(constants.ErrApplicationNotFound)

type ApplicationRepository interface {
	Create(context.Context, *models.Application) error
	FindByID(context.Context, string) (models.Application, error)
	UpdateStatus(context.Context, string, models.ApplicationStatus, string, string, time.Time) error
}

type PostgresApplicationRepository struct {
	db *gorm.DB
}

func NewPostgresApplicationRepository(db *gorm.DB) *PostgresApplicationRepository {
	return &PostgresApplicationRepository{db: db}
}

func (repository *PostgresApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	return repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Create(application).Error
	})
}

func (repository *PostgresApplicationRepository) FindByID(ctx context.Context, id string) (models.Application, error) {
	var application models.Application
	err := repository.db.WithContext(ctx).Preload("Product").Preload("ReviewChecks").First(&application, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Application{}, ErrApplicationNotFound
	}

	return application, err
}

func (repository *PostgresApplicationRepository) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus, reviewedBy, rejectionReason string, reviewedAt time.Time) error {
	result := repository.db.WithContext(ctx).Model(&models.Application{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           status,
			"reviewed_by":      reviewedBy,
			"rejection_reason": rejectionReason,
			"reviewed_at":      reviewedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrApplicationNotFound
	}
	return nil
}
