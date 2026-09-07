package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

func TestPostgresKnowledgeMetricsRepository(t *testing.T) {
	db := sqliteDB(t)
	repo := NewPostgresKnowledgeMetricsRepository(db)
	ctx := context.Background()

	// Initially empty
	metricsEmpty, err := repo.GetKnowledgeMetrics(ctx)
	if err != nil {
		t.Fatalf("GetKnowledgeMetrics(empty) error = %v", err)
	}
	if metricsEmpty.TotalDocuments != 0 || metricsEmpty.TotalChunks != 0 {
		t.Fatalf("expected 0 docs and chunks, got %+v", metricsEmpty)
	}

	// Seed documents
	now := time.Now().UTC()
	docs := []models.KnowledgeDocument{
		{
			ID:           "doc-1",
			Title:        "Doc 1",
			Slug:         "doc-1",
			Category:     models.KnowledgeCategoryUnderwriting,
			Summary:      "Sum 1",
			Content:      "Content 1",
			ChunkCount:   2,
			Status:       models.IndexingStatusIndexed,
			LastSyncedAt: &now,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           "doc-2",
			Title:        "Doc 2",
			Slug:         "doc-2",
			Category:     models.KnowledgeCategoryProduct,
			Summary:      "Sum 2",
			Content:      "Content 2",
			ChunkCount:   1,
			Status:       models.IndexingStatusDraft,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	for i := range docs {
		if err := db.Create(&docs[i]).Error; err != nil {
			t.Fatalf("seed doc error = %v", err)
		}
	}

	// Seed chunks
	chunks := []models.KnowledgeChunk{
		{ID: "c-1", DocumentID: "doc-1", SourceType: "underwriting", Title: "T1", Content: "C1", ChunkIndex: 0},
		{ID: "c-2", DocumentID: "doc-1", SourceType: "underwriting", Title: "T2", Content: "C2", ChunkIndex: 1},
		{ID: "c-3", DocumentID: "doc-2", SourceType: "product", Title: "T3", Content: "C3", ChunkIndex: 0},
	}
	for i := range chunks {
		if err := db.Create(&chunks[i]).Error; err != nil {
			t.Fatalf("seed chunk error = %v", err)
		}
	}

	metrics, err := repo.GetKnowledgeMetrics(ctx)
	if err != nil {
		t.Fatalf("GetKnowledgeMetrics error = %v", err)
	}

	if metrics.TotalDocuments != 2 {
		t.Errorf("TotalDocuments = %d, want 2", metrics.TotalDocuments)
	}
	if metrics.TotalChunks != 3 {
		t.Errorf("TotalChunks = %d, want 3", metrics.TotalChunks)
	}
	if metrics.IndexedDocuments != 1 {
		t.Errorf("IndexedDocuments = %d, want 1", metrics.IndexedDocuments)
	}
	if metrics.DraftDocuments != 1 {
		t.Errorf("DraftDocuments = %d, want 1", metrics.DraftDocuments)
	}
	if metrics.AverageChunksPerDoc != 1.5 {
		t.Errorf("AverageChunksPerDoc = %f, want 1.5", metrics.AverageChunksPerDoc)
	}
	if metrics.CategoryBreakdown[string(models.KnowledgeCategoryUnderwriting)] != 1 {
		t.Errorf("CategoryBreakdown[underwriting] = %d, want 1", metrics.CategoryBreakdown[string(models.KnowledgeCategoryUnderwriting)])
	}
	if metrics.CategoryBreakdown[string(models.KnowledgeCategoryProduct)] != 1 {
		t.Errorf("CategoryBreakdown[product] = %d, want 1", metrics.CategoryBreakdown[string(models.KnowledgeCategoryProduct)])
	}
	if metrics.LastSyncTime == nil {
		t.Errorf("LastSyncTime is nil, want non-nil")
	}
}
