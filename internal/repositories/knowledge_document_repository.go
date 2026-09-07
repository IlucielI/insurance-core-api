package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"gorm.io/gorm"
)

type KnowledgeDocumentRepository interface {
	FindAll(ctx context.Context, category, status, search string) ([]models.KnowledgeDocument, error)
	FindByID(ctx context.Context, id string) (models.KnowledgeDocument, error)
	FindBySlug(ctx context.Context, slug string) (models.KnowledgeDocument, error)
	Create(ctx context.Context, doc *models.KnowledgeDocument) error
	Update(ctx context.Context, id string, doc *models.KnowledgeDocument) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type PostgresKnowledgeDocumentRepository struct {
	db *gorm.DB
}

func NewPostgresKnowledgeDocumentRepository(db *gorm.DB) *PostgresKnowledgeDocumentRepository {
	return &PostgresKnowledgeDocumentRepository{db: db}
}

func (r *PostgresKnowledgeDocumentRepository) FindAll(ctx context.Context, category, status, search string) ([]models.KnowledgeDocument, error) {
	query := r.db.WithContext(ctx).Model(&models.KnowledgeDocument{})

	if category != "" && category != "all" {
		query = query.Where("category = ?", category)
	}

	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(title) LIKE ? OR LOWER(summary) LIKE ? OR LOWER(content) LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	var documents []models.KnowledgeDocument
	if err := query.Order("updated_at DESC").Find(&documents).Error; err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *PostgresKnowledgeDocumentRepository) FindByID(ctx context.Context, id string) (models.KnowledgeDocument, error) {
	var doc models.KnowledgeDocument
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&doc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.KnowledgeDocument{}, errors.New(constants.ErrKnowledgeDocNotFound)
		}
		return models.KnowledgeDocument{}, err
	}
	return doc, nil
}

func (r *PostgresKnowledgeDocumentRepository) FindBySlug(ctx context.Context, slug string) (models.KnowledgeDocument, error) {
	var doc models.KnowledgeDocument
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&doc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.KnowledgeDocument{}, errors.New(constants.ErrKnowledgeDocNotFound)
		}
		return models.KnowledgeDocument{}, err
	}
	return doc, nil
}

func (r *PostgresKnowledgeDocumentRepository) Create(ctx context.Context, doc *models.KnowledgeDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *PostgresKnowledgeDocumentRepository) Update(ctx context.Context, id string, doc *models.KnowledgeDocument) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.KnowledgeDocument
		if err := tx.Where("id = ?", id).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(constants.ErrKnowledgeDocNotFound)
			}
			return err
		}

		return tx.Model(&existing).Select("*").Omit("id", "created_at").Updates(doc).Error
	})
}

func (r *PostgresKnowledgeDocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var doc models.KnowledgeDocument
		if err := tx.Where("id = ?", id).First(&doc).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New(constants.ErrKnowledgeDocNotFound)
			}
			return err
		}

		// Cascade delete any chunks linked to this document
		if err := tx.Where("document_id = ?", id).Delete(&models.KnowledgeChunk{}).Error; err != nil {
			return err
		}

		return tx.Delete(&doc).Error
	})
}

func (r *PostgresKnowledgeDocumentRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.KnowledgeDocument{}).Count(&total).Error
	return total, err
}
