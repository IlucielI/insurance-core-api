package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/google/uuid"
)

type AuditLogService interface {
	Record(ctx context.Context, req dtos.CreateAuditLogRequest) (*dtos.AuditLogResponse, error)
	List(ctx context.Context, query dtos.AuditLogQuery) (*dtos.AuditLogListResponse, error)
	GetByID(ctx context.Context, id string) (*dtos.AuditLogResponse, error)
}

type DefaultAuditLogService struct {
	repo repositories.AuditLogRepository
}

func NewAuditLogService(repo repositories.AuditLogRepository) *DefaultAuditLogService {
	return &DefaultAuditLogService{repo: repo}
}

func (s *DefaultAuditLogService) Record(ctx context.Context, req dtos.CreateAuditLogRequest) (*dtos.AuditLogResponse, error) {
	if err := validations.ValidateCreateAuditLogRequest(&req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("aud_%s_%s", now.Format("20060102150405"), uuid.New().String()[:8])

	detailsJSON, err := validations.DetailsToJSON(req.Details)
	if err != nil {
		return nil, err
	}

	hash := validations.CalculateAuditHash(
		now,
		req.ActorName,
		req.ActorRole,
		req.Action,
		req.Category,
		req.TargetResource,
		req.Status,
		detailsJSON,
	)

	log := models.AuditLog{
		ID:             id,
		Timestamp:      now,
		ActorName:      req.ActorName,
		ActorRole:      req.ActorRole,
		Action:         req.Action,
		Category:       models.AuditCategory(req.Category),
		TargetResource: req.TargetResource,
		IPAddress:      req.IPAddress,
		Status:         models.AuditSeverity(req.Status),
		Details:        detailsJSON,
		Hash:           hash,
	}

	if err := s.repo.Create(ctx, &log); err != nil {
		return nil, err
	}

	res := mapModelToResponse(&log)
	return &res, nil
}

func (s *DefaultAuditLogService) List(ctx context.Context, query dtos.AuditLogQuery) (*dtos.AuditLogListResponse, error) {
	if err := validations.ValidateAuditLogQuery(&query); err != nil {
		return nil, err
	}

	logs, total, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, err
	}

	responses := make([]dtos.AuditLogResponse, 0, len(logs))
	for i := range logs {
		responses = append(responses, mapModelToResponse(&logs[i]))
	}

	return &dtos.AuditLogListResponse{
		Data:  responses,
		Total: total,
	}, nil
}

func (s *DefaultAuditLogService) GetByID(ctx context.Context, id string) (*dtos.AuditLogResponse, error) {
	log, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := mapModelToResponse(log)
	return &res, nil
}

func mapModelToResponse(log *models.AuditLog) dtos.AuditLogResponse {
	details := make(map[string]any)
	if log.Details != "" && log.Details != "{}" {
		if err := json.Unmarshal([]byte(log.Details), &details); err != nil {
			details = make(map[string]any)
		}
	}

	return dtos.AuditLogResponse{
		ID:             log.ID,
		Timestamp:      log.Timestamp.Format(time.RFC3339),
		ActorName:      log.ActorName,
		ActorRole:      log.ActorRole,
		Action:         log.Action,
		Category:       string(log.Category),
		TargetResource: log.TargetResource,
		IPAddress:      log.IPAddress,
		Status:         string(log.Status),
		Details:        details,
		Hash:           log.Hash,
	}
}
