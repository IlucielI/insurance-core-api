package routes

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type routeProductRepository struct{}

func (repository routeProductRepository) FindAll(ctx context.Context, filter repositories.ProductFilter) ([]models.Product, error) {
	return nil, nil
}

func (repository routeProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	return models.Product{}, repositories.ErrProductNotFound
}

type routeApplicationRepository struct{}

func (repository routeApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	return nil
}

func (repository routeApplicationRepository) FindByID(ctx context.Context, id string) (models.Application, error) {
	return models.Application{}, repositories.ErrApplicationNotFound
}

func (repository routeApplicationRepository) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus, reviewedBy, rejectionReason string, reviewedAt time.Time) error {
	return repositories.ErrApplicationNotFound
}

func (repository routeApplicationRepository) List(ctx context.Context, filter repositories.ApplicationListFilter) ([]models.Application, int64, error) {
	return []models.Application{}, 0, nil
}

type routeReviewCheckRepository struct{}

func (repository routeReviewCheckRepository) FindByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationReviewCheck, error) {
	return nil, nil
}

func (repository routeReviewCheckRepository) UpdateStatus(ctx context.Context, applicationID string, checkType models.ApplicationReviewCheckType, status models.ApplicationReviewCheckStatus, reviewedBy, notes string, reviewedAt time.Time) error {
	return repositories.ErrApplicationReviewCheckNotFound
}

func TestNewRouter(t *testing.T) {
	app := NewRouter(config.Config{AppName: "test", Version: "1.0.0", GitHash: "abc123"}, routeProductRepository{}, routeApplicationRepository{}, routeReviewCheckRepository{}, nil, nil)
	if app == nil {
		t.Fatal("NewRouter() = nil")
	}

	request, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}
}
