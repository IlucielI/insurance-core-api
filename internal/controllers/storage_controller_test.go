package controllers

import (
	"context"
	"net/http"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type fakeStorageRepositoryForController struct {
	presignFunc func(ctx context.Context, objectName string, download bool) (string, error)
}

func (repository fakeStorageRepositoryForController) GetUploadURL(ctx context.Context, objectName string) (string, error) {
	return repository.presignFunc(ctx, objectName, false)
}

func (repository fakeStorageRepositoryForController) GetDownloadURL(ctx context.Context, objectName string) (string, error) {
	return repository.presignFunc(ctx, objectName, true)
}

func TestStorageControllerPresign(t *testing.T) {
	controller := NewStorageController(services.NewStorageService(fakeStorageRepositoryForController{presignFunc: func(ctx context.Context, objectName string, download bool) (string, error) {
		if objectName != "documents/file.pdf" || !download {
			t.Fatalf("unexpected args: %s %t", objectName, download)
		}
		return "https://s3.example.com/file", nil
	}}))

	app := fiber.New()
	app.Get("/storage/presign", controller.Presign)

	request, err := http.NewRequest(http.MethodGet, "/storage/presign?object_name=documents/file.pdf&download=true", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
}

func TestStorageControllerReturnsPayload(t *testing.T) {
	controller := NewStorageController(services.NewStorageService(fakeStorageRepositoryForController{presignFunc: func(ctx context.Context, objectName string, download bool) (string, error) {
		return "https://s3.example.com/file", nil
	}}))

	app := fiber.New()
	app.Get("/storage/presign", controller.Presign)

	request, err := http.NewRequest(http.MethodGet, "/storage/presign?object_name=documents/file.pdf", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
}

func TestStorageControllerValidatesInput(t *testing.T) {
	controller := NewStorageController(nil)
	app := fiber.New()
	app.Get("/storage/presign", controller.Presign)

	request, err := http.NewRequest(http.MethodGet, "/storage/presign", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.StatusCode)
	}
}

func TestStorageControllerRejectsMissingObjectName(t *testing.T) {
	controller := NewStorageController(services.NewStorageService(fakeStorageRepositoryForController{presignFunc: func(ctx context.Context, objectName string, download bool) (string, error) {
		return "", nil
	}}))
	app := fiber.New()
	app.Get("/storage/presign", controller.Presign)

	request, err := http.NewRequest(http.MethodGet, "/storage/presign?download=true", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
}

func TestStorageControllerDefaultsInvalidDownloadFlag(t *testing.T) {
	controller := NewStorageController(services.NewStorageService(fakeStorageRepositoryForController{presignFunc: func(ctx context.Context, objectName string, download bool) (string, error) {
		if download {
			t.Fatalf("download flag should be false for invalid value")
		}
		return "https://s3.example.com/file", nil
	}}))
	app := fiber.New()
	app.Get("/storage/presign", controller.Presign)

	request, err := http.NewRequest(http.MethodGet, "/storage/presign?object_name=documents/file.pdf&download=maybe", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
}
