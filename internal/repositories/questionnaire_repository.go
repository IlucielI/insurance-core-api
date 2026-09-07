package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"gorm.io/gorm"
)

var (
	ErrQuestionnaireNotFound = errors.New("questionnaire not found")
)

type QuestionnaireRepository interface {
	FindByProductIDOrCategory(ctx context.Context, productID string, category string) (*models.Questionnaire, []models.Question, error)
	FindQuestionsByQuestionnaireID(ctx context.Context, questionnaireID string) ([]models.Question, error)
	SaveAnswers(ctx context.Context, answers []models.ApplicationAnswer) error
	FindAnswersByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationAnswer, error)
}

type PostgresQuestionnaireRepository struct {
	db    *gorm.DB
	cache ports.Cache
}

func NewPostgresQuestionnaireRepository(db *gorm.DB, cache ...ports.Cache) *PostgresQuestionnaireRepository {
	var c ports.Cache
	if len(cache) > 0 {
		c = cache[0]
	}
	return &PostgresQuestionnaireRepository{db: db, cache: c}
}

func (r *PostgresQuestionnaireRepository) WithCache(cache ports.Cache) *PostgresQuestionnaireRepository {
	r.cache = cache
	return r
}

func (r *PostgresQuestionnaireRepository) FindByProductIDOrCategory(ctx context.Context, productID string, category string) (*models.Questionnaire, []models.Question, error) {
	cacheKey := fmt.Sprintf("tenant:%s:questionnaire:p:%s:c:%s", defaultTenantScope, sanitizeCacheSegment(productID), sanitizeCacheSegment(category))
	if r.cache != nil {
		type cachedResult struct {
			Questionnaire models.Questionnaire `json:"questionnaire"`
			Questions     []models.Question    `json:"questions"`
		}
		var cached cachedResult
		if err := r.cache.GetJSON(ctx, cacheKey, &cached); err == nil && cached.Questionnaire.ID != "" {
			return &cached.Questionnaire, cached.Questions, nil
		}
	}

	var questionnaire models.Questionnaire
	found := false

	// 1. Check by product_id
	if productID != "" {
		err := r.db.WithContext(ctx).
			Where("product_id = ? AND is_active = ?", productID, true).
			Order("version DESC, created_at DESC").
			First(&questionnaire).Error
		if err == nil {
			found = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
	}

	// 2. Check by category
	if !found && category != "" && category != "all" {
		err := r.db.WithContext(ctx).
			Where("category = ? AND is_active = ?", category, true).
			Order("version DESC, created_at DESC").
			First(&questionnaire).Error
		if err == nil {
			found = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, err
		}
	}

	// 3. Fallback to global/all template
	if !found {
		err := r.db.WithContext(ctx).
			Where("is_active = ? AND (product_id IS NULL OR category = 'all')", true).
			Order("version DESC, created_at DESC").
			First(&questionnaire).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrQuestionnaireNotFound
		}
		if err != nil {
			return nil, nil, err
		}
	}

	questions, err := r.FindQuestionsByQuestionnaireID(ctx, questionnaire.ID)
	if err != nil {
		return nil, nil, err
	}

	if r.cache != nil {
		toCache := struct {
			Questionnaire models.Questionnaire `json:"questionnaire"`
			Questions     []models.Question    `json:"questions"`
		}{
			Questionnaire: questionnaire,
			Questions:     questions,
		}
		if err := r.cache.SetJSON(ctx, cacheKey, toCache, 1*time.Hour); err != nil {
			log.Printf("[QuestionnaireRepository] failed to cache questionnaire %s: %v", questionnaire.ID, err)
		}
	}

	return &questionnaire, questions, nil
}

func (r *PostgresQuestionnaireRepository) FindQuestionsByQuestionnaireID(ctx context.Context, questionnaireID string) ([]models.Question, error) {
	var questions []models.Question
	err := r.db.WithContext(ctx).
		Where("questionnaire_id = ? AND is_active = ?", strings.TrimSpace(questionnaireID), true).
		Order("step_number ASC, order_index ASC").
		Find(&questions).Error
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *PostgresQuestionnaireRepository) SaveAnswers(ctx context.Context, answers []models.ApplicationAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&answers).Error
}

func (r *PostgresQuestionnaireRepository) FindAnswersByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationAnswer, error) {
	var answers []models.ApplicationAnswer
	err := r.db.WithContext(ctx).
		Where("application_id = ?", strings.TrimSpace(applicationID)).
		Order("created_at ASC").
		Find(&answers).Error
	if err != nil {
		return nil, err
	}
	return answers, nil
}
