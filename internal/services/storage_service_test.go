package services

import (
	"context"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type fakeStorageRepository struct {
	uploadFunc   func(ctx context.Context, objectName string) (string, error)
	downloadFunc func(ctx context.Context, objectName string) (string, error)
}

func (repository fakeStorageRepository) GetUploadURL(ctx context.Context, objectName string) (string, error) {
	return repository.uploadFunc(ctx, objectName)
}

func (repository fakeStorageRepository) GetDownloadURL(ctx context.Context, objectName string) (string, error) {
	return repository.downloadFunc(ctx, objectName)
}

func TestStorageServiceGetPresignedURL(t *testing.T) {
	service := NewStorageService(fakeStorageRepository{
		uploadFunc: func(ctx context.Context, objectName string) (string, error) {
			if objectName != "documents/file.pdf" {
				t.Fatalf("unexpected object name: %s", objectName)
			}
			return "upload-url", nil
		},
		downloadFunc: func(ctx context.Context, objectName string) (string, error) {
			return "download-url", nil
		},
	})

	link, err := service.GetPresignedURL(context.Background(), "documents/file.pdf", false)
	if err != nil {
		t.Fatalf("GetPresignedURL() error = %v", err)
	}
	if link != "upload-url" {
		t.Fatalf("GetPresignedURL() = %q", link)
	}

	link, err = service.GetPresignedURL(context.Background(), "documents/file.pdf", true)
	if err != nil {
		t.Fatalf("GetPresignedURL(download) error = %v", err)
	}
	if link != "download-url" {
		t.Fatalf("GetPresignedURL(download) = %q", link)
	}
}

func TestStorageServiceValidatesInput(t *testing.T) {
	service := NewStorageService(nil)
	if _, err := service.GetUploadURL(context.Background(), " "); err == nil {
		t.Fatal("GetUploadURL() error = nil, want validation error")
	}
	if _, err := service.GetDownloadURL(context.Background(), " "); err == nil {
		t.Fatal("GetDownloadURL() error = nil, want validation error")
	}
	if _, err := service.GetPresignedURL(context.Background(), "file", true); err == nil {
		t.Fatal("GetPresignedURL() error = nil, want error")
	}
}

var _ repositories.StorageRepository = fakeStorageRepository{}
