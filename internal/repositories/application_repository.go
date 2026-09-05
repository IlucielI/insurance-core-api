package repositories

import (
	"context"
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

var ErrApplicationNotFound = errors.New(constants.ErrApplicationNotFound)

type ApplicationRepository interface {
	Create(context.Context, *models.Application) error
	FindByID(context.Context, string) (models.Application, error)
}

type PostgresApplicationRepository struct {
	db *gorm.DB
}

func NewPostgresApplicationRepository(db *gorm.DB) *PostgresApplicationRepository {
	return &PostgresApplicationRepository{db: db}
}

func (repository *PostgresApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	return repository.db.WithContext(ctx).Create(application).Error
}

func (repository *PostgresApplicationRepository) FindByID(ctx context.Context, id string) (models.Application, error) {
	var application models.Application
	err := repository.db.WithContext(ctx).Preload("Product").First(&application, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Application{}, ErrApplicationNotFound
	}

	return application, err
}
