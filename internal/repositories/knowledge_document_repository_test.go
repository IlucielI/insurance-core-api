package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

func TestPostgresKnowledgeDocumentRepository(t *testing.T) {
	db := sqliteDB(t)
	repo := NewPostgresKnowledgeDocumentRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	doc1 := models.KnowledgeDocument{
		ID:         "doc-1",
		Title:      "Pedoman Underwriting",
		Slug:       "pedoman-underwriting",
		Category:   models.KnowledgeCategoryUnderwriting,
		Summary:    "Ringkasan pedoman underwriting",
		Content:    "Konten detail dokumen underwriting",
		Tags:       []string{"underwriting", "sop"},
		ChunkCount: 5,
		Status:     models.IndexingStatusIndexed,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	doc2 := models.KnowledgeDocument{
		ID:         "doc-2",
		Title:      "Ketentuan Polis Penyakit Kritis",
		Slug:       "ketentuan-polis-penyakit-kritis",
		Category:   models.KnowledgeCategoryProduct,
		Summary:    "Ringkasan klausul penyakit kritis",
		Content:    "Masa tunggu 90 hari penyakit kritis",
		Tags:       []string{"product", "critical-illness"},
		ChunkCount: 8,
		Status:     models.IndexingStatusDraft,
		CreatedAt:  now.Add(-time.Hour),
		UpdatedAt:  now.Add(-time.Hour),
	}

	// 1. Create documents
	if err := repo.Create(ctx, &doc1); err != nil {
		t.Fatalf("Create(doc1) error = %v", err)
	}
	if err := repo.Create(ctx, &doc2); err != nil {
		t.Fatalf("Create(doc2) error = %v", err)
	}

	// 2. Count
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("Count() = %d, want 2", count)
	}

	// 3. FindByID
	found, err := repo.FindByID(ctx, "doc-1")
	if err != nil {
		t.Fatalf("FindByID(doc-1) error = %v", err)
	}
	if found.ID != "doc-1" || found.Title != "Pedoman Underwriting" {
		t.Fatalf("FindByID(doc-1) = %+v, want doc-1", found)
	}

	_, err = repo.FindByID(ctx, "non-existent")
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("FindByID(non-existent) expected ErrKnowledgeDocNotFound, got %v", err)
	}

	// 4. FindBySlug
	foundSlug, err := repo.FindBySlug(ctx, "pedoman-underwriting")
	if err != nil {
		t.Fatalf("FindBySlug() error = %v", err)
	}
	if foundSlug.ID != "doc-1" {
		t.Fatalf("FindBySlug() = %+v, want doc-1", foundSlug)
	}

	_, err = repo.FindBySlug(ctx, "non-existent-slug")
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("FindBySlug(non-existent-slug) expected ErrKnowledgeDocNotFound, got %v", err)
	}

	// 5. FindAll with filters
	// 5a. All
	all, err := repo.FindAll(ctx, "", "", "")
	if err != nil {
		t.Fatalf("FindAll() error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("FindAll() returned %d docs, want 2", len(all))
	}

	// 5b. Category filter
	underwritingDocs, err := repo.FindAll(ctx, "underwriting", "", "")
	if err != nil {
		t.Fatalf("FindAll(category: underwriting) error = %v", err)
	}
	if len(underwritingDocs) != 1 || underwritingDocs[0].ID != "doc-1" {
		t.Fatalf("FindAll(category: underwriting) = %+v, want doc-1", underwritingDocs)
	}

	// 5c. Status filter
	draftDocs, err := repo.FindAll(ctx, "", "draft", "")
	if err != nil {
		t.Fatalf("FindAll(status: draft) error = %v", err)
	}
	if len(draftDocs) != 1 || draftDocs[0].ID != "doc-2" {
		t.Fatalf("FindAll(status: draft) = %+v, want doc-2", draftDocs)
	}

	// 5d. Search query
	searchResults, err := repo.FindAll(ctx, "", "", "kritis")
	if err != nil {
		t.Fatalf("FindAll(search: kritis) error = %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].ID != "doc-2" {
		t.Fatalf("FindAll(search: kritis) = %+v, want doc-2", searchResults)
	}

	// 6. Update
	updatePayload := models.KnowledgeDocument{
		Title:   "Pedoman Underwriting Terbaru",
		Summary: "Ringkasan diperbarui",
	}
	if err := repo.Update(ctx, "doc-1", &updatePayload); err != nil {
		t.Fatalf("Update(doc-1) error = %v", err)
	}
	updated, err := repo.FindByID(ctx, "doc-1")
	if err != nil {
		t.Fatalf("FindByID(doc-1) after update error = %v", err)
	}
	if updated.Title != "Pedoman Underwriting Terbaru" || updated.Summary != "Ringkasan diperbarui" {
		t.Fatalf("updated doc = %+v, want new title and summary", updated)
	}

	err = repo.Update(ctx, "non-existent", &updatePayload)
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("Update(non-existent) expected ErrKnowledgeDocNotFound, got %v", err)
	}

	// 7. Delete (with linked chunk cascade test)
	chunk := models.KnowledgeChunk{
		ID:         "chunk-1",
		DocumentID: "doc-1",
		SourceType: "underwriting",
		Title:      "Pedoman",
		Content:    "Content",
		ChunkIndex: 0,
	}
	if err := db.Create(&chunk).Error; err != nil {
		t.Fatalf("seed chunk error = %v", err)
	}

	if err := repo.Delete(ctx, "doc-1"); err != nil {
		t.Fatalf("Delete(doc-1) error = %v", err)
	}

	// Verify doc-1 is gone
	_, err = repo.FindByID(ctx, "doc-1")
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("FindByID(doc-1) after delete expected not found, got %v", err)
	}

	// Verify linked chunk is deleted
	var chunkCount int64
	if err := db.Model(&models.KnowledgeChunk{}).Where("document_id = ?", "doc-1").Count(&chunkCount).Error; err != nil {
		t.Fatalf("Count chunks error = %v", err)
	}
	if chunkCount != 0 {
		t.Fatalf("expected 0 chunks for deleted doc-1, got %d", chunkCount)
	}

	// Delete non-existent
	err = repo.Delete(ctx, "non-existent")
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("Delete(non-existent) expected ErrKnowledgeDocNotFound, got %v", err)
	}
}
