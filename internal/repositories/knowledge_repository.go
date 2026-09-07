package repositories

import (
	"context"
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	pgvector "github.com/pgvector/pgvector-go"
	"gorm.io/gorm"
)

type KnowledgeChunkMatch struct {
	models.KnowledgeChunk
	Distance float64 `gorm:"column:distance"`
}

type KnowledgeRepository interface {
	ReplaceAll(context.Context, []models.KnowledgeChunk) error
	ReplaceByDocumentID(context.Context, string, []models.KnowledgeChunk) error
	Count(context.Context) (int64, error)
	CountByDocumentID(context.Context, string) (int64, error)
	FindChunksByDocumentID(context.Context, string) ([]models.KnowledgeChunk, error)
	Search(context.Context, []float32, int) ([]KnowledgeChunkMatch, error)
	SearchWithCategory(context.Context, []float32, string, int) ([]KnowledgeChunkMatch, error)
}

type PostgresKnowledgeRepository struct {
	db *gorm.DB
}

func NewPostgresKnowledgeRepository(db *gorm.DB) *PostgresKnowledgeRepository {
	return &PostgresKnowledgeRepository{db: db}
}

func (repository *PostgresKnowledgeRepository) Count(ctx context.Context) (int64, error) {
	var total int64
	err := repository.db.WithContext(ctx).Model(&models.KnowledgeChunk{}).Count(&total).Error
	return total, err
}

func (repository *PostgresKnowledgeRepository) ReplaceAll(ctx context.Context, chunks []models.KnowledgeChunk) error {
	return repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM knowledge_chunks").Error; err != nil {
			return err
		}

		if len(chunks) == 0 {
			return nil
		}

		return tx.Create(&chunks).Error
	})
}

func (repository *PostgresKnowledgeRepository) ReplaceByDocumentID(ctx context.Context, documentID string, chunks []models.KnowledgeChunk) error {
	return repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("document_id = ?", documentID).Delete(&models.KnowledgeChunk{}).Error; err != nil {
			return err
		}

		if len(chunks) == 0 {
			return nil
		}

		return tx.Create(&chunks).Error
	})
}

func (repository *PostgresKnowledgeRepository) CountByDocumentID(ctx context.Context, documentID string) (int64, error) {
	var total int64
	err := repository.db.WithContext(ctx).Model(&models.KnowledgeChunk{}).Where("document_id = ?", documentID).Count(&total).Error
	return total, err
}

func (repository *PostgresKnowledgeRepository) FindChunksByDocumentID(ctx context.Context, documentID string) ([]models.KnowledgeChunk, error) {
	var chunks []models.KnowledgeChunk
	err := repository.db.WithContext(ctx).Omit("embedding").Where("document_id = ?", documentID).Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}

func (repository *PostgresKnowledgeRepository) Search(ctx context.Context, embedding []float32, limit int) ([]KnowledgeChunkMatch, error) {
	return repository.SearchWithCategory(ctx, embedding, "", limit)
}

func (repository *PostgresKnowledgeRepository) SearchWithCategory(ctx context.Context, embedding []float32, category string, limit int) ([]KnowledgeChunkMatch, error) {
	if limit < 1 {
		return nil, errors.New("search limit must be positive")
	}
	if len(embedding) == 0 {
		return nil, errors.New("embedding is required")
	}
	if len(embedding) != constants.AssistantEmbeddingDimension {
		return nil, errors.New("embedding dimension must be 1024")
	}

	queryEmbedding := pgvector.NewVector(embedding)
	query := repository.db.WithContext(ctx).
		Model(&models.KnowledgeChunk{}).
		Select("knowledge_chunks.*, embedding <=> ? AS distance", queryEmbedding)

	if category != "" && category != "all" {
		query = query.Where("source_type = ?", category)
	}

	var matches []KnowledgeChunkMatch
	if err := query.Order("distance ASC").Limit(limit).Scan(&matches).Error; err != nil {
		return nil, err
	}

	return matches, nil
}
