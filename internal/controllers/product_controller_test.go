package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
)

func TestProductManagement_CreateProduct(t *testing.T) {
	prodRepo := &controllerProductRepository{
		products: []models.Product{
			{ID: "prod-existing", Slug: "existing-slug"},
		},
	}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	// 1. Success Create
	reqBody := dtos.CreateProductRequest{
		Name:             "Jiwa Garda",
		Slug:             "jiwa-garda",
		Category:         "life",
		Status:           "active",
		ShortDescription: "Asuransi jiwa garda terdepan",
		Description:      "Deskripsi lengkap asuransi jiwa garda",
		TargetCustomer:   "Profesional muda",
		MinSumAssured:    50000000,
		MaxSumAssured:    1000000000,
		MinPaymentTerm:   5,
		MaxPaymentTerm:   20,
		StartingPremium:  150000,
		NonMCULimit:      500000000,
		Benefits:         []string{"Santunan Kematian"},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("json.Marshal(reqBody) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	// 2. Duplicate Slug -> 409 Conflict
	dupBody := reqBody
	dupBody.Slug = "existing-slug"
	dupPayload, err := json.Marshal(dupBody)
	if err != nil {
		t.Fatalf("json.Marshal(dupBody) error = %v", err)
	}
	dupReq := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(dupPayload))
	dupReq.Header.Set("Content-Type", "application/json")
	dupResp, err := app.Test(dupReq)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if dupResp.StatusCode != http.StatusConflict {
		t.Fatalf("dupResp.StatusCode = %d, want %d", dupResp.StatusCode, http.StatusConflict)
	}

	// 3. Validation error -> 400 Bad Request
	badBody := reqBody
	badBody.MinSumAssured = 0
	badPayload, err := json.Marshal(badBody)
	if err != nil {
		t.Fatalf("json.Marshal(badBody) error = %v", err)
	}
	badReq := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(badPayload))
	badReq.Header.Set("Content-Type", "application/json")
	badResp, err := app.Test(badReq)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if badResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("badResp.StatusCode = %d, want %d", badResp.StatusCode, http.StatusBadRequest)
	}
}

func TestProductManagement_GetByID(t *testing.T) {
	prodRepo := &controllerProductRepository{
		product: models.Product{ID: "prod-100", Name: "Asuransi Sehat", Slug: "asuransi-sehat"},
	}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/id/prod-100", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestProductManagement_UpdateProduct(t *testing.T) {
	prodRepo := &controllerProductRepository{
		product: models.Product{
			ID:             "prod-update-1",
			Name:           "Old Name",
			Slug:           "old-slug",
			MinSumAssured:  10000000,
			MaxSumAssured:  500000000,
			MinPaymentTerm: 5,
			MaxPaymentTerm: 20,
		},
	}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	newName := "New Name"
	updateReq := dtos.UpdateProductRequest{
		Name: &newName,
	}
	payload, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("json.Marshal(updateReq) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/products/prod-update-1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestProductManagement_UpdateStatusAndToggle(t *testing.T) {
	prodRepo := &controllerProductRepository{
		product: models.Product{
			ID:     "prod-status-1",
			Slug:   "prod-status-1",
			Status: models.ProductStatusDraft,
		},
	}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	// 1. Dispatch status update to active via HTTP PATCH
	statusReq := dtos.UpdateProductStatusRequest{Status: "active"}
	payload, err := json.Marshal(statusReq)
	if err != nil {
		t.Fatalf("json.Marshal(statusReq) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/prod-status-1/status", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// 2. Dispatch status toggle via HTTP POST
	toggleReq := httptest.NewRequest(http.MethodPost, "/api/v1/products/prod-status-1/toggle-status", nil)
	toggleResp, err := app.Test(toggleReq)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if toggleResp.StatusCode != http.StatusOK {
		t.Fatalf("toggleResp.StatusCode = %d, want %d", toggleResp.StatusCode, http.StatusOK)
	}
}

func TestProductManagement_Delete(t *testing.T) {
	prodRepo := &controllerProductRepository{
		product: models.Product{
			ID:   "prod-del-1",
			Slug: "prod-del-1",
		},
		hasApps: true,
	}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	// 1. Delete blocked by existing applications -> 409 Conflict
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/prod-del-1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusConflict)
	}

	// 2. Delete allowed -> 200 OK
	prodRepo.hasApps = false
	reqOK := httptest.NewRequest(http.MethodDelete, "/api/v1/products/prod-del-1", nil)
	respOK, err := app.Test(reqOK)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if respOK.StatusCode != http.StatusOK {
		t.Fatalf("respOK.StatusCode = %d, want %d", respOK.StatusCode, http.StatusOK)
	}
}

func TestProductManagement_GetMetrics(t *testing.T) {
	prodRepo := &controllerProductRepository{
		products: []models.Product{
			{ID: "p1", Status: models.ProductStatusActive},
			{ID: "p2", Status: models.ProductStatusDraft},
		},
	}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/products/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestProductManagement_UpdatePricingRules(t *testing.T) {
	prodRepo := &controllerProductRepository{
		product: models.Product{
			ID:   "prod-rules-1",
			Slug: "prod-rules-1",
		},
	}
	ruleRepo := &controllerPricingRuleRepository{}
	app := routes.NewRouter(config.Config{AppName: "test"}, prodRepo, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, ruleRepo)

	rules := []models.ProductPricingRule{
		{RuleCode: "base_rate", RuleName: "Base Rate", RuleType: "base_rate"},
	}
	payload, err := json.Marshal(rules)
	if err != nil {
		t.Fatalf("json.Marshal(rules) error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/products/prod-rules-1/pricing-rules", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
