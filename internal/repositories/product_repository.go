package repositories

import (
	"context"
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

var ErrProductNotFound = errors.New(constants.ErrProductNotFound)

type ProductRepository interface {
	FindAll(ctx context.Context, filter ProductFilter) ([]models.Product, error)
	FindBySlug(ctx context.Context, slug string) (models.Product, error)
}

type ProductFilter struct {
	Category   string
	IsFeatured *bool
	Limit      int
}

type PostgresProductRepository struct {
	db *gorm.DB
}

func NewPostgresProductRepository(db *gorm.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db: db}
}

func (repository *PostgresProductRepository) FindAll(ctx context.Context, filter ProductFilter) ([]models.Product, error) {
	query := repository.db.WithContext(ctx).Order("created_at ASC")

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	if filter.IsFeatured != nil {
		query = query.Where("is_featured = ?", *filter.IsFeatured)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var products []models.Product
	if err := query.Find(&products).Error; err != nil {
		return nil, err
	}

	return products, nil
}

func (repository *PostgresProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	var product models.Product
	err := repository.db.WithContext(ctx).Where("slug = ?", slug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Product{}, ErrProductNotFound
	}
	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}
