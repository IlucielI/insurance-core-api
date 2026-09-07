package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
)

type controllerKnowledgeDocRepository struct {
	docs []models.KnowledgeDocument
	err  error
}

func (r *controllerKnowledgeDocRepository) FindAll(ctx context.Context, category, status, search string) ([]models.KnowledgeDocument, error) {
	if r.err != nil {
		return nil, r.err
	}
	var res []models.KnowledgeDocument
	for _, d := range r.docs {
		if category != "" && string(d.Category) != category {
			continue
		}
		if status != "" && string(d.Status) != status {
			continue
		}
		res = append(res, d)
	}
	return res, nil
}

func (r *controllerKnowledgeDocRepository) FindByID(ctx context.Context, id string) (models.KnowledgeDocument, error) {
	if r.err != nil {
		return models.KnowledgeDocument{}, r.err
	}
	for _, d := range r.docs {
		if d.ID == id {
			return d, nil
		}
	}
	return models.KnowledgeDocument{}, errors.New(constants.ErrKnowledgeDocNotFound)
}

func (r *controllerKnowledgeDocRepository) FindBySlug(ctx context.Context, slug string) (models.KnowledgeDocument, error) {
	if r.err != nil {
		return models.KnowledgeDocument{}, r.err
	}
	for _, d := range r.docs {
		if d.Slug == slug {
			return d, nil
		}
	}
	return models.KnowledgeDocument{}, errors.New(constants.ErrKnowledgeDocNotFound)
}

func (r *controllerKnowledgeDocRepository) Create(ctx context.Context, doc *models.KnowledgeDocument) error {
	if r.err != nil {
		return r.err
	}
	for _, d := range r.docs {
		if d.Slug == doc.Slug {
			return errors.New(constants.ErrKnowledgeDocSlugAlreadyExists)
		}
	}
	r.docs = append(r.docs, *doc)
	return nil
}

func (r *controllerKnowledgeDocRepository) Update(ctx context.Context, id string, doc *models.KnowledgeDocument) error {
	if r.err != nil {
		return r.err
	}
	for i, d := range r.docs {
		if d.ID == id {
			r.docs[i] = *doc
			return nil
		}
	}
	return errors.New(constants.ErrKnowledgeDocNotFound)
}

func (r *controllerKnowledgeDocRepository) Delete(ctx context.Context, id string) error {
	if r.err != nil {
		return r.err
	}
	for i, d := range r.docs {
		if d.ID == id {
			r.docs = append(r.docs[:i], r.docs[i+1:]...)
			return nil
		}
	}
	return errors.New(constants.ErrKnowledgeDocNotFound)
}

func (r *controllerKnowledgeDocRepository) Count(ctx context.Context) (int64, error) {
	if r.err != nil {
		return 0, r.err
	}
	return int64(len(r.docs)), nil
}

func TestKnowledgeDocumentController_List(t *testing.T) {
	repo := &controllerKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "SOP Underwriting", Category: models.KnowledgeCategoryUnderwriting, Status: models.IndexingStatusIndexed},
			{ID: "doc-2", Title: "Klausul Produk", Category: models.KnowledgeCategoryProduct, Status: models.IndexingStatusDraft},
		},
	}

	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, repo)

	// 1. List all
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/documents", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want 200", resp.StatusCode)
	}

	var res struct {
		Data []dtos.KnowledgeDocumentResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		t.Fatalf("json.Decode error = %v", err)
	}
	if len(res.Data) != 2 {
		t.Fatalf("len(res.Data) = %d, want 2", len(res.Data))
	}

	// 2. Filter category
	reqFilter := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/documents?category=underwriting", nil)
	respFilter, err := app.Test(reqFilter)
	if err != nil {
		t.Fatalf("app.Test filter error = %v", err)
	}
	var resFilter struct {
		Data []dtos.KnowledgeDocumentResponse `json:"data"`
	}
	if err := json.NewDecoder(respFilter.Body).Decode(&resFilter); err != nil {
		t.Fatalf("json.Decode filter error = %v", err)
	}
	if len(resFilter.Data) != 1 || resFilter.Data[0].ID != "doc-1" {
		t.Fatalf("filtered count = %d, want 1 (doc-1)", len(resFilter.Data))
	}
}

func TestKnowledgeDocumentController_Get(t *testing.T) {
	repo := &controllerKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "SOP Underwriting", Slug: "sop-underwriting", Category: models.KnowledgeCategoryUnderwriting},
		},
	}

	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, repo)

	// Get by ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/documents/doc-1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	// Get by Slug via slug endpoint
	reqSlug := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/documents/slug/sop-underwriting", nil)
	respSlug, err := app.Test(reqSlug)
	if err != nil {
		t.Fatalf("app.Test slug error = %v", err)
	}
	if respSlug.StatusCode != http.StatusOK {
		t.Fatalf("slug status = %d, want 200", respSlug.StatusCode)
	}

	// Not found
	reqMissing := httptest.NewRequest(http.MethodGet, "/api/v1/knowledge/documents/missing-doc", nil)
	respMissing, err := app.Test(reqMissing)
	if err != nil {
		t.Fatalf("app.Test missing error = %v", err)
	}
	if respMissing.StatusCode != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", respMissing.StatusCode)
	}
}

func TestKnowledgeDocumentController_Create(t *testing.T) {
	repo := &controllerKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-existing", Slug: "existing-slug"},
		},
	}

	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, repo)

	// 1. Success create
	reqBody := dtos.CreateKnowledgeDocumentRequest{
		Title:    "SOP Medis Baru",
		Category: "underwriting",
		Summary:  "Ringkasan SOP pemeriksaan medis",
		Content:  "Konten lengkap teks SOP pemeriksaan medis...",
		Tags:     []string{"medis", "sop"},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("json.Marshal reqBody error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/documents", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("resp.StatusCode = %d, want 201", resp.StatusCode)
	}

	// 2. Conflict on duplicate slug
	reqConflictBody := dtos.CreateKnowledgeDocumentRequest{
		Title:    "Existing Document",
		Slug:     "existing-slug",
		Category: "underwriting",
		Summary:  "Summary",
		Content:  "Content",
	}
	conflictPayload, err := json.Marshal(reqConflictBody)
	if err != nil {
		t.Fatalf("json.Marshal conflict error = %v", err)
	}
	reqConflict := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/documents", bytes.NewReader(conflictPayload))
	reqConflict.Header.Set("Content-Type", "application/json")
	respConflict, err := app.Test(reqConflict)
	if err != nil {
		t.Fatalf("app.Test conflict error = %v", err)
	}
	if respConflict.StatusCode != http.StatusConflict {
		t.Fatalf("respConflict.StatusCode = %d, want 409", respConflict.StatusCode)
	}

	// 3. Validation error on bad body
	reqBad := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/documents", bytes.NewReader([]byte("{invalid-json")))
	reqBad.Header.Set("Content-Type", "application/json")
	respBad, err := app.Test(reqBad)
	if err != nil {
		t.Fatalf("app.Test bad error = %v", err)
	}
	if respBad.StatusCode != http.StatusBadRequest {
		t.Fatalf("respBad.StatusCode = %d, want 400", respBad.StatusCode)
	}
}

func TestKnowledgeDocumentController_Update(t *testing.T) {
	now := time.Now().UTC()
	repo := &controllerKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "Doc One", Slug: "doc-one", Category: models.KnowledgeCategoryUnderwriting, Summary: "Summary", Content: "Content", CreatedAt: now, UpdatedAt: now},
			{ID: "doc-2", Title: "Doc Two", Slug: "doc-two"},
		},
	}

	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, repo)

	// 1. Successful HTTP PUT request
	newTitle := "Updated Doc One"
	updateReq := dtos.UpdateKnowledgeDocumentRequest{
		Title: &newTitle,
	}
	payload, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("json.Marshal updateReq error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/knowledge/documents/doc-1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want 200", resp.StatusCode)
	}

	// 2. Not found
	reqNotFound := httptest.NewRequest(http.MethodPut, "/api/v1/knowledge/documents/doc-99", bytes.NewReader(payload))
	reqNotFound.Header.Set("Content-Type", "application/json")
	respNotFound, err := app.Test(reqNotFound)
	if err != nil {
		t.Fatalf("app.Test not found error = %v", err)
	}
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Fatalf("respNotFound.StatusCode = %d, want 404", respNotFound.StatusCode)
	}

	// 3. Slug conflict
	conflictSlug := "doc-two"
	conflictReq := dtos.UpdateKnowledgeDocumentRequest{
		Slug: &conflictSlug,
	}
	conflictPayload, err := json.Marshal(conflictReq)
	if err != nil {
		t.Fatalf("json.Marshal conflictReq error = %v", err)
	}
	reqConflict := httptest.NewRequest(http.MethodPut, "/api/v1/knowledge/documents/doc-1", bytes.NewReader(conflictPayload))
	reqConflict.Header.Set("Content-Type", "application/json")
	respConflict, err := app.Test(reqConflict)
	if err != nil {
		t.Fatalf("app.Test conflict error = %v", err)
	}
	if respConflict.StatusCode != http.StatusConflict {
		t.Fatalf("respConflict.StatusCode = %d, want 409", respConflict.StatusCode)
	}
}

func TestKnowledgeDocumentController_Delete(t *testing.T) {
	repo := &controllerKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "Doc One"},
		},
	}

	app := routes.NewRouter(config.Config{AppName: "test"}, &controllerProductRepository{}, &controllerApplicationRepository{}, &controllerReviewCheckRepository{}, nil, nil, nil, nil, repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/knowledge/documents/doc-1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want 200", resp.StatusCode)
	}

	reqNotFound := httptest.NewRequest(http.MethodDelete, "/api/v1/knowledge/documents/doc-1", nil)
	respNotFound, err := app.Test(reqNotFound)
	if err != nil {
		t.Fatalf("app.Test not found error = %v", err)
	}
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Fatalf("respNotFound.StatusCode = %d, want 404", respNotFound.StatusCode)
	}
}
