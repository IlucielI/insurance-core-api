package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
)

var nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)

type KnowledgeDocumentService interface {
	GetDocuments(ctx context.Context, category, status, search string) ([]dtos.KnowledgeDocumentResponse, error)
	GetDocumentByID(ctx context.Context, id string) (dtos.KnowledgeDocumentResponse, error)
	GetDocumentBySlug(ctx context.Context, slug string) (dtos.KnowledgeDocumentResponse, error)
	CreateDocument(ctx context.Context, req dtos.CreateKnowledgeDocumentRequest) (dtos.KnowledgeDocumentResponse, error)
	UpdateDocument(ctx context.Context, id string, req dtos.UpdateKnowledgeDocumentRequest) (dtos.KnowledgeDocumentResponse, error)
	DeleteDocument(ctx context.Context, id string) error
}

type DefaultKnowledgeDocumentService struct {
	repo repositories.KnowledgeDocumentRepository
}

func NewKnowledgeDocumentService(repo repositories.KnowledgeDocumentRepository) *DefaultKnowledgeDocumentService {
	return &DefaultKnowledgeDocumentService{repo: repo}
}

func (s *DefaultKnowledgeDocumentService) GetDocuments(ctx context.Context, category, status, search string) ([]dtos.KnowledgeDocumentResponse, error) {
	cat, st, err := validations.ValidateKnowledgeListQuery(category, status)
	if err != nil {
		return nil, err
	}

	docs, err := s.repo.FindAll(ctx, cat, st, strings.TrimSpace(search))
	if err != nil {
		return nil, err
	}

	return dtos.ToKnowledgeDocumentResponses(docs), nil
}

func (s *DefaultKnowledgeDocumentService) GetDocumentByID(ctx context.Context, id string) (dtos.KnowledgeDocumentResponse, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return dtos.KnowledgeDocumentResponse{}, errors.New(constants.ErrKnowledgeDocIDRequired)
	}

	doc, err := s.repo.FindByID(ctx, trimmedID)
	if err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	return dtos.ToKnowledgeDocumentResponse(doc), nil
}

func (s *DefaultKnowledgeDocumentService) GetDocumentBySlug(ctx context.Context, slug string) (dtos.KnowledgeDocumentResponse, error) {
	trimmedSlug := strings.TrimSpace(slug)
	if trimmedSlug == "" {
		return dtos.KnowledgeDocumentResponse{}, errors.New(constants.ErrKnowledgeDocSlugInvalid)
	}

	doc, err := s.repo.FindBySlug(ctx, trimmedSlug)
	if err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	return dtos.ToKnowledgeDocumentResponse(doc), nil
}

func (s *DefaultKnowledgeDocumentService) CreateDocument(ctx context.Context, req dtos.CreateKnowledgeDocumentRequest) (dtos.KnowledgeDocumentResponse, error) {
	if err := validations.ValidateCreateKnowledgeDocRequest(&req); err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	slug := req.Slug
	if slug == "" {
		slug = strings.ToLower(req.Title)
		slug = nonAlphaNumRegex.ReplaceAllString(slug, "-")
		slug = strings.Trim(slug, "-")
		if slug == "" {
			slug = fmt.Sprintf("doc-%d", time.Now().Unix())
		}
	}

	// Check if slug already exists
	if _, err := s.repo.FindBySlug(ctx, slug); err == nil {
		return dtos.KnowledgeDocumentResponse{}, errors.New(constants.ErrKnowledgeDocSlugAlreadyExists)
	}

	// Generate clean deterministic ID from slug
	cleanID := fmt.Sprintf("doc_%s", strings.ReplaceAll(slug, "-", "_"))
	if _, err := s.repo.FindByID(ctx, cleanID); err == nil {
		cleanID = fmt.Sprintf("%s_%d", cleanID, time.Now().Unix())
	}

	estimatedChunks := int(math.Max(2, math.Ceil(float64(len(req.Content))/150.0)))
	now := time.Now().UTC()

	doc := models.KnowledgeDocument{
		ID:           cleanID,
		Title:        req.Title,
		Slug:         slug,
		Category:     models.KnowledgeCategory(req.Category),
		Summary:      req.Summary,
		Content:      req.Content,
		Tags:         req.Tags,
		ChunkCount:   estimatedChunks,
		Status:       models.IndexingStatus(req.Status),
		LastSyncedAt: &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, &doc); err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	return dtos.ToKnowledgeDocumentResponse(doc), nil
}

func (s *DefaultKnowledgeDocumentService) UpdateDocument(ctx context.Context, id string, req dtos.UpdateKnowledgeDocumentRequest) (dtos.KnowledgeDocumentResponse, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return dtos.KnowledgeDocumentResponse{}, errors.New(constants.ErrKnowledgeDocIDRequired)
	}

	if err := validations.ValidateUpdateKnowledgeDocRequest(&req); err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	existing, err := s.repo.FindByID(ctx, trimmedID)
	if err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	if req.Slug != nil && *req.Slug != existing.Slug {
		if conflict, err := s.repo.FindBySlug(ctx, *req.Slug); err == nil && conflict.ID != existing.ID {
			return dtos.KnowledgeDocumentResponse{}, errors.New(constants.ErrKnowledgeDocSlugAlreadyExists)
		}
		existing.Slug = *req.Slug
	}

	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Category != nil {
		existing.Category = models.KnowledgeCategory(*req.Category)
	}
	if req.Summary != nil {
		existing.Summary = *req.Summary
	}
	if req.Content != nil {
		existing.Content = *req.Content
		existing.ChunkCount = int(math.Max(2, math.Ceil(float64(len(existing.Content))/150.0)))
	}
	if req.Tags != nil {
		existing.Tags = *req.Tags
	}
	if req.Status != nil {
		existing.Status = models.IndexingStatus(*req.Status)
	}

	now := time.Now().UTC()
	existing.UpdatedAt = now

	if err := s.repo.Update(ctx, existing.ID, &existing); err != nil {
		return dtos.KnowledgeDocumentResponse{}, err
	}

	return dtos.ToKnowledgeDocumentResponse(existing), nil
}

func (s *DefaultKnowledgeDocumentService) DeleteDocument(ctx context.Context, id string) error {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return errors.New(constants.ErrKnowledgeDocIDRequired)
	}

	return s.repo.Delete(ctx, trimmedID)
}
