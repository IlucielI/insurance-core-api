package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultConversationCacheTTL = 24 * time.Hour
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
)

type AssistantConversationRepository interface {
	GetOrCreateConversation(ctx context.Context, id string, title string) (models.AssistantConversation, error)
	GetConversation(ctx context.Context, id string) (models.AssistantConversation, error)
	ListMessages(ctx context.Context, conversationID string, limit int) ([]models.AssistantMessage, error)
	SaveMessage(ctx context.Context, msg models.AssistantMessage) error
	SaveMessages(ctx context.Context, msgs []models.AssistantMessage) error
	DeleteConversation(ctx context.Context, id string) error
}

type PostgresAssistantConversationRepository struct {
	db    *gorm.DB
	cache ports.Cache
}

func NewPostgresAssistantConversationRepository(db *gorm.DB, cache ...ports.Cache) *PostgresAssistantConversationRepository {
	var c ports.Cache
	if len(cache) > 0 {
		c = cache[0]
	}
	return &PostgresAssistantConversationRepository{db: db, cache: c}
}

func (r *PostgresAssistantConversationRepository) WithCache(cache ports.Cache) *PostgresAssistantConversationRepository {
	r.cache = cache
	return r
}

func (r *PostgresAssistantConversationRepository) GetOrCreateConversation(ctx context.Context, id string, title string) (models.AssistantConversation, error) {
	conv := models.AssistantConversation{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{"updated_at": time.Now()}),
		}).
		Create(&conv).Error
	if err != nil {
		return models.AssistantConversation{}, err
	}

	return r.GetConversation(ctx, id)
}

func (r *PostgresAssistantConversationRepository) GetConversation(ctx context.Context, id string) (models.AssistantConversation, error) {
	var conv models.AssistantConversation
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&conv, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.AssistantConversation{}, ErrConversationNotFound
		}
		return models.AssistantConversation{}, err
	}

	return conv, nil
}

func (r *PostgresAssistantConversationRepository) ListMessages(ctx context.Context, conversationID string, limit int) ([]models.AssistantMessage, error) {
	cacheKey := fmt.Sprintf("assistant:conversation:%s:messages", conversationID)
	if r.cache != nil {
		var cached []models.AssistantMessage
		if err := r.cache.GetJSON(ctx, cacheKey, &cached); err == nil && len(cached) > 0 {
			if limit > 0 && len(cached) > limit {
				return cached[len(cached)-limit:], nil
			}
			return cached, nil
		}
	}

	var messages []models.AssistantMessage
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	if r.cache != nil && len(messages) > 0 {
		_ = r.cache.SetJSON(ctx, cacheKey, messages, defaultConversationCacheTTL)
	}

	if limit > 0 && len(messages) > limit {
		return messages[len(messages)-limit:], nil
	}

	return messages, nil
}

func (r *PostgresAssistantConversationRepository) SaveMessage(ctx context.Context, msg models.AssistantMessage) error {
	return r.SaveMessages(ctx, []models.AssistantMessage{msg})
}

func (r *PostgresAssistantConversationRepository) SaveMessages(ctx context.Context, msgs []models.AssistantMessage) error {
	if len(msgs) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Create(&msgs).Error; err != nil {
		return err
	}

	if r.cache != nil && len(msgs) > 0 {
		cacheKey := fmt.Sprintf("assistant:conversation:%s:messages", msgs[0].ConversationID)
		_ = r.cache.Delete(ctx, cacheKey)
	}

	return nil
}

func (r *PostgresAssistantConversationRepository) DeleteConversation(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&models.AssistantConversation{}, "id = ?", id).Error; err != nil {
		return err
	}

	if r.cache != nil {
		cacheKey := fmt.Sprintf("assistant:conversation:%s:messages", id)
		_ = r.cache.Delete(ctx, cacheKey)
	}

	return nil
}
