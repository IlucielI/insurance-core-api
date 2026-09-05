package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
	"github.com/gofiber/fiber/v2"
)

type controllerProductRepository struct {
	products []models.Product
	product  models.Product
	err      error
}

func (repository *controllerProductRepository) FindAll(ctx context.Context, filter repositories.ProductFilter) ([]models.Product, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	return repository.products, nil
}

func (repository *controllerProductRepository) FindBySlug(ctx context.Context, slug string) (models.Product, error) {
	if repository.err != nil {
		return models.Product{}, repository.err
	}
	return repository.product, nil
}

type controllerApplicationRepository struct {
	application models.Application
	err         error
	created     *models.Application
}

func (repository *controllerApplicationRepository) Create(ctx context.Context, application *models.Application) error {
	repository.created = application
	return repository.err
}

func (repository *controllerApplicationRepository) FindByID(ctx context.Context, id string) (models.Application, error) {
	if repository.err != nil {
		return models.Application{}, repository.err
	}
	return repository.application, nil
}

func (repository *controllerApplicationRepository) UpdateStatus(ctx context.Context, id string, status models.ApplicationStatus, reviewedBy, rejectionReason string, reviewedAt time.Time) error {
	return repository.err
}

func TestHealthController(t *testing.T) {
	app := routes.NewRouter(config.Config{AppName: "insurance-core-api", Version: "1.2.3", GitHash: "abc123"}, &controllerProductRepository{}, &controllerApplicationRepository{})

	response := performRequest(t, app, http.MethodGet, "/health", nil)
	assertStatus(t, response, http.StatusOK)
	body := readBody(t, response)
	assertBodyContains(t, body, "\"version\":\"1.2.3\"")
	assertBodyContains(t, body, "\"git_hash\":\"abc123\"")
	assertBodyContains(t, body, "\"uptime\":")
}

func TestProductRoutes(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{products: []models.Product{product}, product: product}, &controllerApplicationRepository{})

	response := performRequest(t, app, http.MethodGet, "/api/v1/products?category=life&featured=true&limit=1", nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, readBody(t, response), "Secure Life Plus")

	response = performRequest(t, app, http.MethodGet, "/api/v1/products/secure-life-plus", nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, readBody(t, response), "secure-life-plus")

	response = performRequest(t, app, http.MethodPost, "/api/v1/products/secure-life-plus/quotes", productQuoteBody())
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, readBody(t, response), "estimated_premium")
}

func TestProductRoutesHandleErrors(t *testing.T) {
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{err: repositories.ErrProductNotFound}, &controllerApplicationRepository{})

	response := performRequest(t, app, http.MethodGet, "/api/v1/products?category=travel", nil)
	assertStatus(t, response, http.StatusBadRequest)
	assertBodyContains(t, readBody(t, response), constants.ErrProductCategoryInvalid)

	response = performRequest(t, app, http.MethodGet, "/api/v1/products/missing", nil)
	assertStatus(t, response, http.StatusNotFound)

	response = performRequest(t, app, http.MethodPost, "/api/v1/products/missing/quotes", productQuoteBody())
	assertStatus(t, response, http.StatusNotFound)

	response = performRequest(t, app, http.MethodPost, "/api/v1/products/missing/quotes", bytes.NewBufferString("{"))
	assertStatus(t, response, http.StatusBadRequest)
	assertBodyContains(t, readBody(t, response), constants.ErrProductQuoteBodyInvalid)
}

func TestApplicationRoutes(t *testing.T) {
	product := controllerProductFixture()
	application := models.Application{ID: "application-1", ProductID: product.ID, Product: product, Status: models.ApplicationStatusSubmitted}
	applications := &controllerApplicationRepository{application: application}
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, applications)

	response := performRequest(t, app, http.MethodPost, "/api/v1/products/secure-life-plus/applications", applicationBody())
	assertStatus(t, response, http.StatusCreated)
	if applications.created == nil || applications.created.Status != models.ApplicationStatusSubmitted {
		t.Fatalf("created application = %+v, want submitted application", applications.created)
	}

	response = performRequest(t, app, http.MethodGet, "/api/v1/applications/application-1", nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, readBody(t, response), "application-1")

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/status", jsonBody(map[string]any{
		"status":      models.ApplicationStatusUnderReview,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusNoContent)
}

func TestApplicationRoutesHandleErrors(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, &controllerApplicationRepository{err: repositories.ErrApplicationNotFound})

	response := performRequest(t, app, http.MethodPost, "/api/v1/products/secure-life-plus/applications", bytes.NewBufferString("{"))
	assertStatus(t, response, http.StatusBadRequest)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationBodyInvalid)

	response = performRequest(t, app, http.MethodPost, "/api/v1/products/secure-life-plus/applications", jsonBody(map[string]any{"full_name": "x"}))
	assertStatus(t, response, http.StatusBadRequest)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationFullNameInvalid)

	response = performRequest(t, app, http.MethodGet, "/api/v1/applications/missing", nil)
	assertStatus(t, response, http.StatusNotFound)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/missing/status", jsonBody(map[string]any{
		"status":      models.ApplicationStatusUnderReview,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusNotFound)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/missing/status", bytes.NewBufferString("{"))
	assertStatus(t, response, http.StatusBadRequest)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/missing/status", jsonBody(map[string]any{"status": models.ApplicationStatusUnderReview}))
	assertStatus(t, response, http.StatusBadRequest)
}

func TestApplicationRouteMapsInternalErrors(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, &controllerApplicationRepository{application: models.Application{ID: "application-1", Status: models.ApplicationStatusSubmitted}, err: errors.New("db failed")})

	response := performRequest(t, app, http.MethodGet, "/api/v1/applications/application-1", nil)
	assertStatus(t, response, http.StatusInternalServerError)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationGetFailed)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/status", jsonBody(map[string]any{
		"status":      models.ApplicationStatusUnderReview,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusInternalServerError)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationStatusUpdateFailed)
}

func performRequest(t *testing.T, app *fiber.App, method string, path string, body io.Reader) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test() error = %v", err)
	}
	return response
}

func readBody(t *testing.T, response *http.Response) string {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return string(body)
}

func assertStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("status = %d, want %d", response.StatusCode, want)
	}
}

func assertBodyContains(t *testing.T, body string, value string) {
	t.Helper()
	if !bytes.Contains([]byte(body), []byte(value)) {
		t.Fatalf("body = %s, want contains %q", body, value)
	}
}

func jsonBody(value any) io.Reader {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(body)
}

func productQuoteBody() io.Reader {
	return jsonBody(map[string]any{
		"age":               35,
		"gender":            constants.GenderMale,
		"sum_assured":       300_000_000,
		"payment_term":      10,
		"payment_frequency": constants.PaymentFrequencyMonthly,
		"smoker":            constants.SmokerNo,
		"occupation_class":  constants.OccupationStandard,
		"health_risk":       constants.HealthRiskLow,
	})
}

func applicationBody() io.Reader {
	return jsonBody(map[string]any{
		"full_name":         "Bayu Anugerah",
		"email":             "bayu@example.com",
		"phone":             "+628123456789",
		"age":               35,
		"gender":            constants.GenderMale,
		"sum_assured":       300_000_000,
		"payment_term":      10,
		"payment_frequency": constants.PaymentFrequencyMonthly,
		"smoker":            constants.SmokerNo,
		"occupation_class":  constants.OccupationStandard,
		"health_risk":       constants.HealthRiskLow,
	})
}

func controllerProductFixture() models.Product {
	return models.Product{
		ID:               "product-1",
		Name:             "Secure Life Plus",
		Slug:             "secure-life-plus",
		Category:         models.ProductCategoryLife,
		ShortDescription: "Life protection",
		Description:      "Life protection",
		TargetCustomer:   "Families",
		MinSumAssured:    100_000_000,
		MaxSumAssured:    2_000_000_000,
		MinPaymentTerm:   5,
		MaxPaymentTerm:   30,
		StartingPremium:  100_000,
		PricingRules: models.PricingRules{
			BaseRate: 0.004,
			AgeFactors: []models.AgeFactor{
				{MinAge: 18, MaxAge: 30, Factor: 1.0},
				{MinAge: 31, MaxAge: 45, Factor: 1.2},
			},
			GenderFactors: map[string]float64{
				constants.GenderMale:   1.0,
				constants.GenderFemale: 0.95,
			},
			SmokerFactors: map[string]float64{
				constants.SmokerNo:  1.0,
				constants.SmokerYes: 1.4,
			},
			OccupationFactors: map[string]float64{
				constants.OccupationLow:      0.95,
				constants.OccupationStandard: 1.0,
				constants.OccupationHigh:     1.25,
			},
			HealthFactors: map[string]float64{
				constants.HealthRiskLow:    1.0,
				constants.HealthRiskMedium: 1.15,
				constants.HealthRiskHigh:   1.5,
			},
			FrequencyLoading: map[string]float64{
				constants.PaymentFrequencyAnnual:     1.0,
				constants.PaymentFrequencySemiAnnual: 1.03,
				constants.PaymentFrequencyQuarterly:  1.06,
				constants.PaymentFrequencyMonthly:    1.10,
			},
		},
	}
}
