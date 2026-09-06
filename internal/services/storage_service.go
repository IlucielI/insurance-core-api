package services

import (
	"context"
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type StorageService struct {
	repository repositories.StorageRepository
}

func NewStorageService(repository repositories.StorageRepository) *StorageService {
	return &StorageService{repository: repository}
}

func (service *StorageService) GetUploadURL(ctx context.Context, objectName string) (string, error) {
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return "", errors.New("object name is required")
	}
	if service.repository == nil {
		return "", errors.New("storage service unavailable")
	}

	return service.repository.GetUploadURL(ctx, objectName)
}

func (service *StorageService) GetPresignedURL(ctx context.Context, objectName string, download bool) (string, error) {
	if download {
		return service.GetDownloadURL(ctx, objectName)
	}

	return service.GetUploadURL(ctx, objectName)
}

func (service *StorageService) GetDownloadURL(ctx context.Context, objectName string) (string, error) {
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return "", errors.New("object name is required")
	}
	if service.repository == nil {
		return "", errors.New("storage service unavailable")
	}

	return service.repository.GetDownloadURL(ctx, objectName)
}
