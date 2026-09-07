package repositories

import (
	"context"
	"testing"
	"time"
)

type fakeObjectStorage struct {
	putFunc func(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
	getFunc func(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error)
}

func (storage fakeObjectStorage) PresignPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	return storage.putFunc(ctx, bucketName, objectName, expiry)
}

func (storage fakeObjectStorage) PresignGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	return storage.getFunc(ctx, bucketName, objectName, expiry)
}

func TestS3StorageRepositoryGetUploadURL(t *testing.T) {
	repository, err := NewS3StorageRepository(fakeObjectStorage{
		putFunc: func(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
			if bucketName != "insurance-files" || objectName != "documents/file.pdf" || expiry != 15*time.Minute {
				t.Fatalf("unexpected args: %s %s %s", bucketName, objectName, expiry)
			}
			return "https://s3.example.com/upload", nil
		},
	}, "insurance-files", 0, 0, "")
	if err != nil {
		t.Fatalf("NewS3StorageRepository() error = %v", err)
	}

	link, err := repository.GetUploadURL(context.Background(), "documents/file.pdf")
	if err != nil {
		t.Fatalf("GetUploadURL() error = %v", err)
	}
	if link != "https://s3.example.com/upload" {
		t.Fatalf("GetUploadURL() = %q", link)
	}
}

func TestS3StorageRepositoryGetDownloadURLOverride(t *testing.T) {
	repository, err := NewS3StorageRepository(fakeObjectStorage{}, "insurance-files", 0, 0, "https://cdn.example.com/files")
	if err != nil {
		t.Fatalf("NewS3StorageRepository() error = %v", err)
	}

	link, err := repository.GetDownloadURL(context.Background(), "documents/file.pdf")
	if err != nil {
		t.Fatalf("GetDownloadURL() error = %v", err)
	}
	if link != "https://cdn.example.com/files/documents/file.pdf" {
		t.Fatalf("GetDownloadURL() = %q", link)
	}
}

func TestS3StorageRepositoryGetDownloadURLOverrideEncodesObjectName(t *testing.T) {
	repository, err := NewS3StorageRepository(fakeObjectStorage{}, "insurance-files", 0, 0, "https://cdn.example.com/files")
	if err != nil {
		t.Fatalf("NewS3StorageRepository() error = %v", err)
	}

	link, err := repository.GetDownloadURL(context.Background(), "documents/file baru.pdf")
	if err != nil {
		t.Fatalf("GetDownloadURL() error = %v", err)
	}
	if link != "https://cdn.example.com/files/documents/file%20baru.pdf" {
		t.Fatalf("GetDownloadURL() = %q", link)
	}
}

func TestS3StorageRepositoryRejectsTraversalObjectName(t *testing.T) {
	repository, err := NewS3StorageRepository(fakeObjectStorage{}, "insurance-files", 0, 0, "https://cdn.example.com/files")
	if err != nil {
		t.Fatalf("NewS3StorageRepository() error = %v", err)
	}

	invalidObjectNames := []string{
		"../private/secret.txt",
		"..",
		"documents/../secret.txt",
		"documents\\..\\secret.txt",
	}

	for _, objectName := range invalidObjectNames {
		if _, err := repository.GetDownloadURL(context.Background(), objectName); err == nil || err.Error() != "invalid object name" {
			t.Fatalf("GetDownloadURL(%q) error = %v, want invalid object name", objectName, err)
		}
	}
}

func TestS3StorageRepositoryValidatesInput(t *testing.T) {
	repository, err := NewS3StorageRepository(fakeObjectStorage{
		putFunc: func(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
			return "upload-url", nil
		},
		getFunc: func(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
			return "download-url", nil
		},
	}, "insurance-files", 0, 0, "")
	if err != nil {
		t.Fatalf("NewS3StorageRepository() error = %v", err)
	}
	if _, err := repository.GetUploadURL(context.Background(), " "); err == nil {
		t.Fatal("GetUploadURL() error = nil, want validation error")
	}
	if _, err := repository.GetDownloadURL(context.Background(), " "); err == nil {
		t.Fatal("GetDownloadURL() error = nil, want validation error")
	}
	if _, err := repository.GetUploadURL(context.Background(), "../file"); err == nil || err.Error() != "invalid object name" {
		t.Fatalf("GetUploadURL('../file') error = %v, want 'invalid object name'", err)
	}
}

func TestS3StorageRepositoryConstructorValidatesInput(t *testing.T) {
	if _, err := NewS3StorageRepository(nil, "bucket", 0, 0, ""); err == nil {
		t.Fatal("NewS3StorageRepository(nil, ...) error = nil, want error")
	}
	if _, err := NewS3StorageRepository(fakeObjectStorage{}, " ", 0, 0, ""); err == nil {
		t.Fatal("NewS3StorageRepository(..., empty bucket) error = nil, want error")
	}
}
