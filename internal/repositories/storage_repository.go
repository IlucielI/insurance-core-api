package repositories

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/ports"
)

type StorageRepository interface {
	GetUploadURL(ctx context.Context, objectName string) (string, error)
	GetDownloadURL(ctx context.Context, objectName string) (string, error)
}

type S3StorageRepository struct {
	storage         ports.ObjectStorage
	bucketName      string
	uploadExpiry    time.Duration
	downloadExpiry  time.Duration
	overrideBaseURL string
}

func NewS3StorageRepository(storage ports.ObjectStorage, bucketName string, uploadExpiry, downloadExpiry time.Duration, overrideBaseURL string) (*S3StorageRepository, error) {
	if storage == nil {
		return nil, errors.New("storage client is required")
	}
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return nil, errors.New("bucket name is required")
	}

	return &S3StorageRepository{
		storage:         storage,
		bucketName:      bucketName,
		uploadExpiry:    normalizeStorageExpiry(uploadExpiry, 15*time.Minute),
		downloadExpiry:  normalizeStorageExpiry(downloadExpiry, 24*time.Hour),
		overrideBaseURL: strings.TrimRight(strings.TrimSpace(overrideBaseURL), "/"),
	}, nil
}

func (repository *S3StorageRepository) GetUploadURL(ctx context.Context, objectName string) (string, error) {
	cleanObjectName, err := sanitizeObjectName(objectName)
	if err != nil {
		return "", err
	}
	if repository.storage == nil {
		return "", errors.New("storage client is required")
	}
	if repository.bucketName == "" {
		return "", errors.New("bucket name is required")
	}
	return repository.storage.PresignPutObject(ctx, repository.bucketName, cleanObjectName, repository.uploadExpiry)
}

func (repository *S3StorageRepository) GetDownloadURL(ctx context.Context, objectName string) (string, error) {
	cleanObjectName, err := sanitizeObjectName(objectName)
	if err != nil {
		return "", err
	}
	if repository.storage == nil {
		return "", errors.New("storage client is required")
	}
	if repository.bucketName == "" {
		return "", errors.New("bucket name is required")
	}
	if repository.overrideBaseURL != "" {
		joinedURL, err := url.JoinPath(repository.overrideBaseURL, cleanObjectName)
		if err != nil {
			return "", fmt.Errorf("failed to join override base url: %w", err)
		}
		return joinedURL, nil
	}
	return repository.storage.PresignGetObject(ctx, repository.bucketName, cleanObjectName, repository.downloadExpiry)
}

func sanitizeObjectName(objectName string) (string, error) {
	trimmedObjectName := strings.TrimSpace(objectName)
	if trimmedObjectName == "" {
		return "", errors.New("object name is required")
	}
	if strings.Contains(trimmedObjectName, "\\") {
		return "", errors.New("invalid object name")
	}
	for _, pathSegment := range strings.Split(trimmedObjectName, "/") {
		if pathSegment == ".." {
			return "", errors.New("invalid object name")
		}
	}

	cleanObjectName := path.Clean(strings.TrimLeft(trimmedObjectName, "/"))
	if cleanObjectName == "." {
		return "", errors.New("invalid object name")
	}
	return cleanObjectName, nil
}

func normalizeStorageExpiry(expiry, fallback time.Duration) time.Duration {
	if expiry <= 0 {
		return fallback
	}
	return expiry
}
