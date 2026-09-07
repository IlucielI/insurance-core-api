package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
	"github.com/gofiber/fiber/v2"
)

type controllerProductRepository struct {
	products []models.Product
	product  models.Product
	err      error
	hasApps  bool
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
	if repository.product.Slug == slug || repository.product.ID == slug {
		return repository.product, nil
	}
	for _, p := range repository.products {
		if p.Slug == slug || p.ID == slug {
			return p, nil
		}
	}
	return models.Product{}, repositories.ErrProductNotFound
}

func (repository *controllerProductRepository) FindByID(ctx context.Context, id string) (models.Product, error) {
	if repository.err != nil {
		return models.Product{}, repository.err
	}
	if repository.product.ID == id || repository.product.Slug == id {
		return repository.product, nil
	}
	for _, p := range repository.products {
		if p.ID == id || p.Slug == id {
			return p, nil
		}
	}
	return models.Product{}, repositories.ErrProductNotFound
}

func (repository *controllerProductRepository) Create(ctx context.Context, product *models.Product) error {
	if repository.err != nil {
		return repository.err
	}
	repository.products = append(repository.products, *product)
	repository.product = *product
	return nil
}

func (repository *controllerProductRepository) Update(ctx context.Context, product *models.Product) error {
	if repository.err != nil {
		return repository.err
	}
	repository.product = *product
	return nil
}

func (repository *controllerProductRepository) UpdateStatus(ctx context.Context, id string, status models.ProductStatus) error {
	if repository.err != nil {
		return repository.err
	}
	repository.product.Status = status
	return nil
}

func (repository *controllerProductRepository) Delete(ctx context.Context, id string) error {
	return repository.err
}

func (repository *controllerProductRepository) GetMetrics(ctx context.Context) (dtos.ProductManagementMetricsResponse, error) {
	if repository.err != nil {
		return dtos.ProductManagementMetricsResponse{}, repository.err
	}
	return dtos.ProductManagementMetricsResponse{
		TotalProducts:  len(repository.products),
		ActiveProducts: len(repository.products),
	}, nil
}

func (repository *controllerProductRepository) HasApplications(ctx context.Context, productID string) (bool, error) {
	if repository.err != nil {
		return false, repository.err
	}
	return repository.hasApps, nil
}

type controllerPricingRuleRepository struct {
	rules []models.ProductPricingRule
	err   error
}

func (r *controllerPricingRuleRepository) SaveBatch(ctx context.Context, productID string, rules []models.ProductPricingRule) error {
	if r.err != nil {
		return r.err
	}
	r.rules = rules
	return nil
}

func (r *controllerPricingRuleRepository) FindByProductID(ctx context.Context, productID string) ([]models.ProductPricingRule, error) {
	return r.rules, r.err
}

func (r *controllerPricingRuleRepository) FindByProductSlug(ctx context.Context, slug string) ([]models.ProductPricingRule, error) {
	return r.rules, r.err
}

func (r *controllerPricingRuleRepository) Create(ctx context.Context, rule *models.ProductPricingRule) error {
	r.rules = append(r.rules, *rule)
	return r.err
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

func (repository *controllerApplicationRepository) List(ctx context.Context, filter repositories.ApplicationListFilter) ([]models.Application, int64, error) {
	if repository.err != nil {
		return nil, 0, repository.err
	}
	return []models.Application{repository.application}, 1, nil
}

type controllerReviewCheckRepository struct {
	checks []models.ApplicationReviewCheck
	err    error
}

func (repository *controllerReviewCheckRepository) FindByApplicationID(ctx context.Context, applicationID string) ([]models.ApplicationReviewCheck, error) {
	if repository.err != nil {
		return nil, repository.err
	}
	return repository.checks, nil
}

func (repository *controllerReviewCheckRepository) UpdateStatus(ctx context.Context, applicationID string, checkType models.ApplicationReviewCheckType, status models.ApplicationReviewCheckStatus, reviewedBy, notes string, reviewedAt time.Time) error {
	return repository.err
}

func TestHealthController(t *testing.T) {
	app := routes.NewRouter(config.Config{AppName: "insurance-core-api", Version: "1.2.3", GitHash: "abc123"}, &controllerProductRepository{}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	response := performRequest(t, app, http.MethodGet, "/health", nil)
	assertStatus(t, response, http.StatusOK)
	body := readBody(t, response)
	assertBodyContains(t, body, "\"version\":\"1.2.3\"")
	assertBodyContains(t, body, "\"git_hash\":\"abc123\"")
	assertBodyContains(t, body, "\"uptime\":")
}

func TestProductRoutes(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{products: []models.Product{product}, product: product}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	response := performRequest(t, app, http.MethodGet, "/api/v1/products?category=life&featured=true&limit=1&search=secure", nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, readBody(t, response), "Secure Life Plus")

	response = performRequest(t, app, http.MethodGet, "/api/v1/products/secure-life-plus", nil)
	assertStatus(t, response, http.StatusOK)
	detailBody := readBody(t, response)
	assertBodyContains(t, detailBody, "secure-life-plus")
	assertBodyContains(t, detailBody, "sum_assured_presets")

	response = performRequest(t, app, http.MethodPost, "/api/v1/products/secure-life-plus/quotes", productQuoteBody())
	assertStatus(t, response, http.StatusOK)
	body := readBody(t, response)
	assertBodyContains(t, body, "estimated_premium")
	assertBodyContains(t, body, "\"product_slug\":\"secure-life-plus\"")

	// Test GET /api/v1/products/:slug/pricing-rules
	mockPricingRepo := &controllerPricingRuleRepository{
		rules: []models.ProductPricingRule{
			{ID: "rule-1", ProductID: product.ID, RuleCode: "base_rate", RuleName: "Base Rate", RuleType: "base_rate"},
		},
	}
	appWithPricing := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{products: []models.Product{product}, product: product}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, mockPricingRepo)
	response = performRequest(t, appWithPricing, http.MethodGet, "/api/v1/products/secure-life-plus/pricing-rules", nil)
	assertStatus(t, response, http.StatusOK)
	pricingBody := readBody(t, response)
	assertBodyContains(t, pricingBody, "rule-1")
	assertBodyContains(t, pricingBody, "base_rate")
}

func TestProductRoutesHandleErrors(t *testing.T) {
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{err: repositories.ErrProductNotFound}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	response := performRequest(t, app, http.MethodGet, "/api/v1/products?category=travel", nil)
	assertStatus(t, response, http.StatusBadRequest)
	assertBodyContains(t, readBody(t, response), constants.ErrProductCategoryInvalid)

	longSearch := strings.Repeat("a", 101)
	response = performRequest(t, app, http.MethodGet, "/api/v1/products?search="+longSearch, nil)
	assertStatus(t, response, http.StatusBadRequest)
	assertBodyContains(t, readBody(t, response), constants.ErrProductSearchInvalid)

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
	reviewChecks := &controllerReviewCheckRepository{checks: []models.ApplicationReviewCheck{{ID: "check-1", ApplicationID: application.ID, CheckType: models.ApplicationReviewCheckTypeIdentityVerified, Status: models.ApplicationReviewCheckStatusPending}}}
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, applications, reviewChecks, nil, nil, nil, nil)

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

	response = performRequest(t, app, http.MethodGet, "/api/v1/applications/application-1/review-checks", nil)
	assertStatus(t, response, http.StatusOK)
	assertBodyContains(t, readBody(t, response), string(models.ApplicationReviewCheckTypeIdentityVerified))

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/review-checks/identity_verified", jsonBody(map[string]any{
		"status":      models.ApplicationReviewCheckStatusPassed,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusNoContent)
}

func TestApplicationRoutesHandleErrors(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, &controllerApplicationRepository{err: repositories.ErrApplicationNotFound}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

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

func TestApplicationReviewCheckRoutesHandleErrors(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, &controllerApplicationRepository{application: models.Application{ID: "application-1"}}, &controllerReviewCheckRepository{err: repositories.ErrApplicationReviewCheckNotFound}, nil, nil, nil, nil)

	response := performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/review-checks/identity_verified", jsonBody(map[string]any{
		"status":      models.ApplicationReviewCheckStatusPassed,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusNotFound)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/review-checks/identity_verified", bytes.NewBufferString("{"))
	assertStatus(t, response, http.StatusBadRequest)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/review-checks/unknown", jsonBody(map[string]any{
		"status":      models.ApplicationReviewCheckStatusPassed,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusBadRequest)
}

func TestApplicationRouteMapsInternalErrors(t *testing.T) {
	product := controllerProductFixture()
	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, &controllerApplicationRepository{application: models.Application{ID: "application-1", Status: models.ApplicationStatusSubmitted}, err: errors.New("db failed")}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	response := performRequest(t, app, http.MethodGet, "/api/v1/applications/application-1", nil)
	assertStatus(t, response, http.StatusInternalServerError)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationGetFailed)

	response = performRequest(t, app, http.MethodPatch, "/api/v1/applications/application-1/status", jsonBody(map[string]any{
		"status":      models.ApplicationStatusUnderReview,
		"reviewed_by": "underwriter",
	}))
	assertStatus(t, response, http.StatusInternalServerError)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationStatusUpdateFailed)

	app = routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{product: product}, &controllerApplicationRepository{application: models.Application{ID: "application-1"}}, &controllerReviewCheckRepository{err: errors.New("db failed")}, nil, nil, nil, nil)
	response = performRequest(t, app, http.MethodGet, "/api/v1/applications/application-1/review-checks", nil)
	assertStatus(t, response, http.StatusInternalServerError)
	assertBodyContains(t, readBody(t, response), constants.ErrApplicationGetFailed)
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
			SumAssuredPresets:  []int64{100_000_000, 250_000_000, 500_000_000, 1_000_000_000},
			PaymentTermPresets: []int{5, 10, 15, 20},
		},
	}
}
