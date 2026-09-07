package services

import (
	"context"
	"fmt"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/google/uuid"
)

type NotificationService interface {
	List(ctx context.Context, query dtos.NotificationQuery) (*dtos.NotificationListResponse, error)
	Create(ctx context.Context, req dtos.CreateNotificationRequest) (*dtos.NotificationResponse, error)
	CreateBatch(ctx context.Context, reqs []dtos.CreateNotificationRequest) ([]*dtos.NotificationResponse, error)
	MarkAsRead(ctx context.Context, id string) (*dtos.MarkReadResponse, error)
	MarkAllAsRead(ctx context.Context) (*dtos.MarkAllReadResponse, error)
	GetUnreadCount(ctx context.Context) (int64, error)
}

type DefaultNotificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) *DefaultNotificationService {
	return &DefaultNotificationService{repo: repo}
}

func (s *DefaultNotificationService) List(ctx context.Context, query dtos.NotificationQuery) (*dtos.NotificationListResponse, error) {
	if err := validations.ValidateNotificationQuery(&query); err != nil {
		return nil, err
	}

	items, total, unreadCount, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, err
	}

	resData := make([]dtos.NotificationResponse, 0, len(items))
	for _, item := range items {
		resData = append(resData, toNotificationResponse(item))
	}

	return &dtos.NotificationListResponse{
		Data:        resData,
		Total:       total,
		UnreadCount: unreadCount,
	}, nil
}

func (s *DefaultNotificationService) Create(ctx context.Context, req dtos.CreateNotificationRequest) (*dtos.NotificationResponse, error) {
	if err := validations.ValidateCreateNotificationRequest(&req); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	id := fmt.Sprintf("notif_%s_%s", now.Format("20060102150405"), uuid.New().String()[:8])

	item := models.Notification{
		ID:        id,
		Type:      req.Type,
		Category:  models.NotificationCategory(req.Category),
		Severity:  models.NotificationSeverity(req.Severity),
		Title:     req.Title,
		Message:   req.Message,
		Link:      req.Link,
		IsRead:    false,
		CreatedAt: now,
	}

	if err := s.repo.Create(ctx, &item); err != nil {
		return nil, err
	}

	resp := toNotificationResponse(item)
	return &resp, nil
}

func (s *DefaultNotificationService) CreateBatch(ctx context.Context, reqs []dtos.CreateNotificationRequest) ([]*dtos.NotificationResponse, error) {
	if len(reqs) == 0 {
		return []*dtos.NotificationResponse{}, nil
	}

	items := make([]models.Notification, 0, len(reqs))
	responses := make([]*dtos.NotificationResponse, 0, len(reqs))
	now := time.Now().UTC()

	for _, req := range reqs {
		copyReq := req
		if err := validations.ValidateCreateNotificationRequest(&copyReq); err != nil {
			return nil, err
		}

		id := fmt.Sprintf("notif_%s_%s", now.Format("20060102150405"), uuid.New().String()[:8])
		item := models.Notification{
			ID:        id,
			Type:      copyReq.Type,
			Category:  models.NotificationCategory(copyReq.Category),
			Severity:  models.NotificationSeverity(copyReq.Severity),
			Title:     copyReq.Title,
			Message:   copyReq.Message,
			Link:      copyReq.Link,
			IsRead:    false,
			CreatedAt: now,
		}
		items = append(items, item)
		resp := toNotificationResponse(item)
		responses = append(responses, &resp)
	}

	if err := s.repo.CreateBatch(ctx, items); err != nil {
		return nil, err
	}

	return responses, nil
}

func (s *DefaultNotificationService) MarkAsRead(ctx context.Context, id string) (*dtos.MarkReadResponse, error) {
	now := time.Now().UTC()
	item, err := s.repo.MarkAsRead(ctx, id, now)
	if err != nil {
		return nil, err
	}

	readAtStr := ""
	if item.ReadAt != nil {
		readAtStr = item.ReadAt.Format(time.RFC3339)
	}

	return &dtos.MarkReadResponse{
		ID:     item.ID,
		IsRead: item.IsRead,
		ReadAt: readAtStr,
	}, nil
}

func (s *DefaultNotificationService) MarkAllAsRead(ctx context.Context) (*dtos.MarkAllReadResponse, error) {
	now := time.Now().UTC()
	count, err := s.repo.MarkAllAsRead(ctx, now)
	if err != nil {
		return nil, err
	}

	unreadCount, err := s.repo.CountUnread(ctx)
	if err != nil {
		return nil, err
	}

	return &dtos.MarkAllReadResponse{
		UpdatedCount: count,
		UnreadCount:  unreadCount,
	}, nil
}

func (s *DefaultNotificationService) GetUnreadCount(ctx context.Context) (int64, error) {
	return s.repo.CountUnread(ctx)
}

func toNotificationResponse(item models.Notification) dtos.NotificationResponse {
	var readAt *string
	if item.ReadAt != nil {
		str := item.ReadAt.Format(time.RFC3339)
		readAt = &str
	}

	return dtos.NotificationResponse{
		ID:        item.ID,
		Type:      item.Type,
		Category:  string(item.Category),
		Severity:  string(item.Severity),
		Title:     item.Title,
		Message:   item.Message,
		Link:      item.Link,
		IsRead:    item.IsRead,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
		ReadAt:    readAt,
	}
}
