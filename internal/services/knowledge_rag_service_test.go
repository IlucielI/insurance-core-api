package services

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type fakeRAGKnowledgeRepository struct {
	chunks     []models.KnowledgeChunk
	matches    []repositories.KnowledgeChunkMatch
	replaceErr error
	searchErr  error
}

func (f *fakeRAGKnowledgeRepository) ReplaceAll(_ context.Context, c []models.KnowledgeChunk) error {
	f.chunks = c
	return f.replaceErr
}

func (f *fakeRAGKnowledgeRepository) ReplaceByDocumentID(_ context.Context, docID string, c []models.KnowledgeChunk) error {
	if f.replaceErr != nil {
		return f.replaceErr
	}
	var kept []models.KnowledgeChunk
	for _, chunk := range f.chunks {
		if chunk.DocumentID != docID {
			kept = append(kept, chunk)
		}
	}
	f.chunks = append(kept, c...)
	return nil
}

func (f *fakeRAGKnowledgeRepository) Count(_ context.Context) (int64, error) {
	return int64(len(f.chunks)), nil
}

func (f *fakeRAGKnowledgeRepository) CountByDocumentID(_ context.Context, docID string) (int64, error) {
	var count int64
	for _, c := range f.chunks {
		if c.DocumentID == docID {
			count++
		}
	}
	return count, nil
}

func (f *fakeRAGKnowledgeRepository) FindChunksByDocumentID(_ context.Context, docID string) ([]models.KnowledgeChunk, error) {
	var res []models.KnowledgeChunk
	for _, c := range f.chunks {
		if c.DocumentID == docID {
			res = append(res, c)
		}
	}
	return res, nil
}

func (f *fakeRAGKnowledgeRepository) Search(_ context.Context, _ []float32, _ int) ([]repositories.KnowledgeChunkMatch, error) {
	return f.matches, f.searchErr
}

func (f *fakeRAGKnowledgeRepository) SearchWithCategory(_ context.Context, _ []float32, category string, limit int) ([]repositories.KnowledgeChunkMatch, error) {
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	var filtered []repositories.KnowledgeChunkMatch
	for _, m := range f.matches {
		if category != "" && category != "all" && m.SourceType != category {
			continue
		}
		filtered = append(filtered, m)
		if len(filtered) == limit {
			break
		}
	}
	return filtered, nil
}

type fakeMetricsRepository struct {
	metrics dtos.KnowledgeBaseMetricsResponse
	err     error
}

func (f *fakeMetricsRepository) GetKnowledgeMetrics(_ context.Context) (dtos.KnowledgeBaseMetricsResponse, error) {
	if f.err != nil {
		return dtos.KnowledgeBaseMetricsResponse{}, f.err
	}
	return f.metrics, nil
}

type fakeRAGLLM struct {
	embeddingResp []float32
	embeddingErr  error
	chatResp      string
	chatErr       error
}

func (f *fakeRAGLLM) CreateEmbedding(_ context.Context, _ llm.EmbeddingInput) ([]float32, error) {
	if f.embeddingErr != nil {
		return nil, f.embeddingErr
	}
	if len(f.embeddingResp) > 0 {
		return f.embeddingResp, nil
	}
	return generateDeterministicEmbedding("default", 1024), nil
}

func (f *fakeRAGLLM) CreateChatCompletion(_ context.Context, _ llm.ChatCompletionInput) (string, error) {
	if f.chatErr != nil {
		return "", f.chatErr
	}
	return f.chatResp, nil
}

func TestChunkKnowledgeDocument(t *testing.T) {
	// 1. Empty content
	emptyDoc := models.KnowledgeDocument{
		ID:      "doc-empty",
		Content: "   ",
	}
	if chunks := ChunkKnowledgeDocument(emptyDoc); chunks != nil {
		t.Fatalf("expected nil chunks for empty document, got %v", chunks)
	}

	// 2. Normal content
	doc := models.KnowledgeDocument{
		ID:       "doc-1",
		Title:    "Pedoman Underwriting",
		Category: models.KnowledgeCategoryUnderwriting,
		Content:  "Satu dua tiga empat lima enam tujuh delapan sembilan sepuluh",
	}
	chunks := ChunkKnowledgeDocument(doc)
	if len(chunks) == 0 {
		t.Fatal("expected at least 1 chunk")
	}
	if chunks[0].DocumentID != "doc-1" || chunks[0].SourceType != "underwriting" {
		t.Fatalf("chunk metadata mismatch: %+v", chunks[0])
	}
}

func TestDeterministicEmbedding(t *testing.T) {
	vec := generateDeterministicEmbedding("test text", 1024)
	if len(vec) != 1024 {
		t.Fatalf("vector dimension = %d, want 1024", len(vec))
	}

	// Verify L2 normalization
	var sumSq float64
	for _, v := range vec {
		sumSq += float64(v * v)
	}
	norm := math.Sqrt(sumSq)
	if math.Abs(norm-1.0) > 1e-4 {
		t.Fatalf("vector norm = %f, want ~1.0", norm)
	}

	// Deterministic verification
	vec2 := generateDeterministicEmbedding("test text", 1024)
	if vec[0] != vec2[0] || vec[500] != vec2[500] {
		t.Fatal("deterministic embedding produced different values for identical text")
	}
}

func TestKnowledgeRAGService_SyncAndReindex(t *testing.T) {
	now := time.Now().UTC()
	docRepo := &fakeKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{
				ID:        "doc-1",
				Title:     "Doc 1",
				Slug:      "doc-1-slug",
				Category:  models.KnowledgeCategoryUnderwriting,
				Content:   "Pemeriksaan non-medical limit usia muda.",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
	chunkRepo := &fakeRAGKnowledgeRepository{}
	metricsRepo := &fakeMetricsRepository{}
	llmClient := &fakeRAGLLM{}

	service := NewKnowledgeRAGService(docRepo, chunkRepo, metricsRepo, llmClient)

	// Reindex valid document
	resp, err := service.ReindexDocument(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("ReindexDocument error = %v", err)
	}
	if resp.DocumentID != "doc-1" || resp.Status != string(models.IndexingStatusIndexed) {
		t.Fatalf("unexpected reindex response: %+v", resp)
	}
	if len(chunkRepo.chunks) == 0 {
		t.Fatal("expected chunks to be populated after reindex")
	}

	// Reindex by slug
	respSlug, err := service.ReindexDocument(context.Background(), "doc-1-slug")
	if err != nil {
		t.Fatalf("ReindexDocument by slug error = %v", err)
	}
	if respSlug.DocumentID != "doc-1" {
		t.Fatalf("reindex by slug = %s, want doc-1", respSlug.DocumentID)
	}

	// Reindex missing document
	_, err = service.ReindexDocument(context.Background(), "missing-doc")
	if err == nil {
		t.Fatal("expected error for missing document")
	}

	// Reindex empty ID
	_, err = service.ReindexDocument(context.Background(), "")
	if err == nil || err.Error() != constants.ErrKnowledgeDocIDRequired {
		t.Fatalf("expected ErrKnowledgeDocIDRequired, got %v", err)
	}
}

func TestKnowledgeRAGService_GetMetrics(t *testing.T) {
	expectedMetrics := dtos.KnowledgeBaseMetricsResponse{
		TotalDocuments:      5,
		TotalChunks:         50,
		IndexedDocuments:    4,
		SyncingDocuments:    0,
		DraftDocuments:      1,
		AverageChunksPerDoc: 10.0,
	}
	metricsRepo := &fakeMetricsRepository{metrics: expectedMetrics}
	service := NewKnowledgeRAGService(nil, nil, metricsRepo, nil)

	metrics, err := service.GetMetrics(context.Background())
	if err != nil {
		t.Fatalf("GetMetrics error = %v", err)
	}
	if metrics.TotalDocuments != 5 || metrics.TotalChunks != 50 {
		t.Fatalf("metrics mismatch: %+v", metrics)
	}
}

func TestKnowledgeRAGService_SimulateChat(t *testing.T) {
	chunkRepo := &fakeRAGKnowledgeRepository{
		matches: []repositories.KnowledgeChunkMatch{
			{
				KnowledgeChunk: models.KnowledgeChunk{
					ID:         "chk-1",
					DocumentID: "doc-1",
					Title:      "Pedoman Underwriting",
					SourceType: "underwriting",
					Content:    "Non-medical limit usia 18-35 adalah 500 juta rupiah.",
					ChunkIndex: 0,
				},
				Distance: 0.15,
			},
		},
	}
	llmClient := &fakeRAGLLM{
		chatResp: "Berdasarkan pedoman underwriting, limit non-medis adalah Rp 500.000.000.",
	}

	service := NewKnowledgeRAGService(nil, chunkRepo, nil, llmClient)

	// 1. Success query
	resp, err := service.SimulateChat(context.Background(), dtos.SimulateChatRequest{
		Query:    "Berapa limit non medis usia 30 tahun?",
		Category: "underwriting",
		TopK:     3,
	})
	if err != nil {
		t.Fatalf("SimulateChat error = %v", err)
	}
	if resp.Answer != llmClient.chatResp {
		t.Fatalf("SimulateChat answer = %q, want %q", resp.Answer, llmClient.chatResp)
	}
	if len(resp.RetrievedChunks) != 1 {
		t.Fatalf("len(RetrievedChunks) = %d, want 1", len(resp.RetrievedChunks))
	}
	if resp.RetrievedChunks[0].Similarity != 0.85 {
		t.Fatalf("RetrievedChunks similarity = %f, want 0.85", resp.RetrievedChunks[0].Similarity)
	}

	// 2. Empty query validation
	_, err = service.SimulateChat(context.Background(), dtos.SimulateChatRequest{Query: "   "})
	if err == nil || err.Error() != "query is required" {
		t.Fatalf("expected query is required error, got %v", err)
	}

	// 3. Fallback when LLM chat response is empty
	llmEmpty := &fakeRAGLLM{chatResp: ""}
	serviceFallback := NewKnowledgeRAGService(nil, chunkRepo, nil, llmEmpty)
	respFallback, err := serviceFallback.SimulateChat(context.Background(), dtos.SimulateChatRequest{
		Query: "info limit",
	})
	if err != nil {
		t.Fatalf("SimulateChat fallback error = %v", err)
	}
	if respFallback.Answer == "" {
		t.Fatal("expected non-empty fallback answer")
	}

	// 4. No matches found
	emptyChunkRepo := &fakeRAGKnowledgeRepository{}
	serviceNoMatches := NewKnowledgeRAGService(nil, emptyChunkRepo, nil, nil)
	respNoMatches, err := serviceNoMatches.SimulateChat(context.Background(), dtos.SimulateChatRequest{
		Query: "sesuatu yang tidak ada",
	})
	if err != nil {
		t.Fatalf("SimulateChat no matches error = %v", err)
	}
	if respNoMatches.Answer == "" || len(respNoMatches.RetrievedChunks) != 0 {
		t.Fatalf("unexpected response when no matches: %+v", respNoMatches)
	}
}
