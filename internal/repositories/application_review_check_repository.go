package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

var ErrApplicationReviewCheckNotFound = errors.New(constants.ErrApplicationReviewCheckNotFound)

type ApplicationReviewCheckRepository interface {
	FindByApplicationID(context.Context, string) ([]models.ApplicationReviewCheck, error)
	UpdateStatus(context.Context, string, models.ApplicationReviewCheckType, models.ApplicationReviewCheckStatus, string, string, time.Time) error
}

type PostgresApplicationReviewCheckRepository struct {
	db *gorm.DB
}

func NewPostgresApplicationReviewCheckRepository(db *gorm.DB) *PostgresApplicationReviewCheckRepository {
	return &PostgresApplicationReviewCheckRepository{db: db}
}

func (repository *PostgresApplicationReviewCheckRepository) FindByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationReviewCheck, error) {
	var checks []models.ApplicationReviewCheck
	if err := repository.db.WithContext(ctx).Where("application_id = ?", applicationID).Order("created_at ASC").Find(&checks).Error; err != nil {
		return nil, err
	}

	return checks, nil
}

func (repository *PostgresApplicationReviewCheckRepository) UpdateStatus(ctx context.Context, applicationID string, checkType models.ApplicationReviewCheckType, status models.ApplicationReviewCheckStatus, reviewedBy, notes string, reviewedAt time.Time) error {
	result := repository.db.WithContext(ctx).Model(&models.ApplicationReviewCheck{}).
		Where("application_id = ? AND check_type = ?", applicationID, checkType).
		Updates(map[string]interface{}{
			"status":      status,
			"reviewed_by": reviewedBy,
			"notes":       notes,
			"reviewed_at": reviewedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrApplicationReviewCheckNotFound
	}

	return nil
}
