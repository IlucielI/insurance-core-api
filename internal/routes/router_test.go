package routes

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type routeProductRepository struct{}

func (repository routeProductRepository) FindAll(ctx context.Context, filter repositories.ProductFilter) ([]models.Product, error) {
	return nil, nil
}

func (repository routeProductRepository) FindByID(ctx context.Context, id string) (models.Product, error) {
	return models.Product{}, repositories.ErrProductNotFound
}

func (repository routeProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	return models.Product{}, repositories.ErrProductNotFound
}

func (repository routeProductRepository) Create(ctx context.Context, product *models.Product) error {
	return nil
}

func (repository routeProductRepository) Update(ctx context.Context, product *models.Product) error {
	return nil
}

func (repository routeProductRepository) UpdateStatus(ctx context.Context, id string, status models.ProductStatus) error {
	return nil
}

func (repository routeProductRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func (repository routeProductRepository) GetMetrics(ctx context.Context) (dtos.ProductManagementMetricsResponse, error) {
	return dtos.ProductManagementMetricsResponse{}, nil
}

func (repository routeProductRepository) HasApplications(ctx context.Context, productID string) (bool, error) {
	return false, nil
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
	app := NewRouter(config.Config{AppName: "test", Version: "1.0.0", GitHash: "abc123"}, routeProductRepository{}, routeApplicationRepository{}, routeReviewCheckRepository{}, nil, nil, nil, nil)
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

func TestStorageRouteExists(t *testing.T) {
	app := NewRouter(config.Config{AppName: "test"}, routeProductRepository{}, routeApplicationRepository{}, routeReviewCheckRepository{}, nil, nil, nil, nil)
	request, err := http.NewRequest(http.MethodGet, "/api/v1/storage/presign?object_name=a.txt", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.StatusCode)
	}
}

func TestAssistantStreamRouteExists(t *testing.T) {
	app := NewRouter(config.Config{AppName: "test"}, routeProductRepository{}, routeApplicationRepository{}, routeReviewCheckRepository{}, nil, nil, nil, nil)
	request, err := http.NewRequest(http.MethodPost, "/api/v1/assistant/chat/stream", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	// With nil assistantService, it should return 503 Service Unavailable, not 404 Not Found
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", response.StatusCode)
	}
}

type routeMetricsRepository struct{}

func (r routeMetricsRepository) GetAdminMetrics(ctx context.Context) (models.AdminMetrics, error) {
	return models.AdminMetrics{}, nil
}

func TestAdminMetricsRouteExists(t *testing.T) {
	app := NewRouter(config.Config{AppName: "test"}, routeProductRepository{}, routeApplicationRepository{}, routeReviewCheckRepository{}, nil, nil, nil, nil, routeMetricsRepository{})
	request, err := http.NewRequest(http.MethodGet, "/api/v1/admin/metrics", nil)
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

type routeAuditLogRepository struct{}

func (r routeAuditLogRepository) Create(_ context.Context, _ *models.AuditLog) error {
	return nil
}

func (r routeAuditLogRepository) FindAll(_ context.Context, _ dtos.AuditLogQuery) ([]models.AuditLog, int64, error) {
	return []models.AuditLog{}, 0, nil
}

func (r routeAuditLogRepository) FindByID(_ context.Context, _ string) (*models.AuditLog, error) {
	return nil, nil
}

func (r routeAuditLogRepository) Count(_ context.Context) (int64, error) {
	return 0, nil
}

type routeSystemHealthService struct{}

func (r routeSystemHealthService) GetOverview(_ context.Context) (*dtos.SystemHealthOverviewResponse, error) {
	return &dtos.SystemHealthOverviewResponse{}, nil
}

func (r routeSystemHealthService) PingServices(_ context.Context, _ string) ([]dtos.ServiceHealthItem, error) {
	return []dtos.ServiceHealthItem{}, nil
}

func (r routeSystemHealthService) PingRoutes(_ context.Context) ([]dtos.RouteLatencyProbeItem, error) {
	return []dtos.RouteLatencyProbeItem{}, nil
}

func TestAdminHealthAndAuditRoutesExist(t *testing.T) {
	app := NewRouter(
		config.Config{AppName: "test"},
		routeProductRepository{},
		routeApplicationRepository{},
		routeReviewCheckRepository{},
		nil, nil, nil, nil,
		routeAuditLogRepository{},
		routeSystemHealthService{},
	)

	t.Run("health overview route", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/admin/health/overview", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})

	t.Run("audit logs route", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/admin/audit-logs", nil)
		if err != nil {
			t.Fatalf("NewRequest() error = %v", err)
		}
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test() error = %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want 200", resp.StatusCode)
		}
	})
}

