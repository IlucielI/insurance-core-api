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

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/controllers"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/gofiber/fiber/v2"
)

type fakeKnowledgeRAGService struct {
	reindexResp dtos.ReindexDocumentResponse
	reindexErr  error
	metricsResp dtos.KnowledgeBaseMetricsResponse
	metricsErr  error
	chatResp    dtos.SimulateChatResponse
	chatErr     error
}

func (f *fakeKnowledgeRAGService) ReindexDocument(_ context.Context, id string) (dtos.ReindexDocumentResponse, error) {
	if f.reindexErr != nil {
		return dtos.ReindexDocumentResponse{}, f.reindexErr
	}
	if id == "not-found" {
		return dtos.ReindexDocumentResponse{}, errors.New(constants.ErrKnowledgeDocNotFound)
	}
	return f.reindexResp, nil
}

func (f *fakeKnowledgeRAGService) GetMetrics(_ context.Context) (dtos.KnowledgeBaseMetricsResponse, error) {
	if f.metricsErr != nil {
		return dtos.KnowledgeBaseMetricsResponse{}, f.metricsErr
	}
	return f.metricsResp, nil
}

func (f *fakeKnowledgeRAGService) SimulateChat(_ context.Context, req dtos.SimulateChatRequest) (dtos.SimulateChatResponse, error) {
	if f.chatErr != nil {
		return dtos.SimulateChatResponse{}, f.chatErr
	}
	return f.chatResp, nil
}

func (f *fakeKnowledgeRAGService) SyncDocumentChunks(_ context.Context, _ models.KnowledgeDocument) (int, error) {
	return 0, nil
}

func setupRAGApp(service *fakeKnowledgeRAGService) *fiber.App {
	app := fiber.New()
	ctrl := controllers.NewKnowledgeRAGController(service)

	api := app.Group("/api/v1")
	api.Post("/knowledge/documents/:id/reindex", ctrl.Reindex)
	api.Get("/admin/knowledge/metrics", ctrl.GetMetrics)
	api.Post("/knowledge/simulate-chat", ctrl.SimulateChat)

	return app
}

func TestKnowledgeRAGController_Reindex(t *testing.T) {
	service := &fakeKnowledgeRAGService{
		reindexResp: dtos.ReindexDocumentResponse{
			DocumentID:   "doc-1",
			Title:        "Doc 1",
			Status:       "indexed",
			ChunkCount:   5,
			LastSyncedAt: time.Now().UTC(),
		},
	}
	app := setupRAGApp(service)

	// Successful reindex
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/documents/doc-1/reindex", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want 200", resp.StatusCode)
	}

	// Missing document
	reqMissing := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/documents/not-found/reindex", nil)
	respMissing, err := app.Test(reqMissing)
	if err != nil {
		t.Fatalf("app.Test missing error = %v", err)
	}
	if respMissing.StatusCode != http.StatusNotFound {
		t.Fatalf("respMissing.StatusCode = %d, want 404", respMissing.StatusCode)
	}
}

func TestKnowledgeRAGController_GetMetrics(t *testing.T) {
	service := &fakeKnowledgeRAGService{
		metricsResp: dtos.KnowledgeBaseMetricsResponse{
			TotalDocuments: 10,
			TotalChunks:    100,
		},
	}
	app := setupRAGApp(service)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/knowledge/metrics", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want 200", resp.StatusCode)
	}

	var body struct {
		Data dtos.KnowledgeBaseMetricsResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json.Decode error = %v", err)
	}
	if body.Data.TotalDocuments != 10 {
		t.Fatalf("TotalDocuments = %d, want 10", body.Data.TotalDocuments)
	}

	// Internal error scenario
	service.metricsErr = errors.New("db error")
	reqErr := httptest.NewRequest(http.MethodGet, "/api/v1/admin/knowledge/metrics", nil)
	respErr, err := app.Test(reqErr)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if respErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("respErr.StatusCode = %d, want 500", respErr.StatusCode)
	}
}

func TestKnowledgeRAGController_SimulateChat(t *testing.T) {
	service := &fakeKnowledgeRAGService{
		chatResp: dtos.SimulateChatResponse{
			Query:           "pertanyaan limit",
			Answer:          "Limitnya adalah 500 juta.",
			ExecutionTimeMs: 45,
		},
	}
	app := setupRAGApp(service)

	// Successful simulate chat
	chatReq := dtos.SimulateChatRequest{
		Query:    "pertanyaan limit",
		Category: "underwriting",
		TopK:     3,
	}
	payload, err := json.Marshal(chatReq)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/simulate-chat", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode = %d, want 200", resp.StatusCode)
	}

	// Missing query
	badReq := dtos.SimulateChatRequest{
		Query: "   ",
	}
	badPayload, err := json.Marshal(badReq)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}
	reqBad := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/simulate-chat", bytes.NewReader(badPayload))
	reqBad.Header.Set("Content-Type", "application/json")
	respBad, err := app.Test(reqBad)
	if err != nil {
		t.Fatalf("app.Test bad query error = %v", err)
	}
	if respBad.StatusCode != http.StatusBadRequest {
		t.Fatalf("respBad.StatusCode = %d, want 400", respBad.StatusCode)
	}

	// Malformed JSON
	reqMalformed := httptest.NewRequest(http.MethodPost, "/api/v1/knowledge/simulate-chat", bytes.NewReader([]byte("{bad-json")))
	reqMalformed.Header.Set("Content-Type", "application/json")
	respMalformed, err := app.Test(reqMalformed)
	if err != nil {
		t.Fatalf("app.Test malformed error = %v", err)
	}
	if respMalformed.StatusCode != http.StatusBadRequest {
		t.Fatalf("respMalformed.StatusCode = %d, want 400", respMalformed.StatusCode)
	}
}
