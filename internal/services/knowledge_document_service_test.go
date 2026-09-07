package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

type fakeKnowledgeDocRepository struct {
	docs []models.KnowledgeDocument
	err  error
}

func (f *fakeKnowledgeDocRepository) FindAll(ctx context.Context, category, status, search string) ([]models.KnowledgeDocument, error) {
	if f.err != nil {
		return nil, f.err
	}
	var res []models.KnowledgeDocument
	for _, d := range f.docs {
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

func (f *fakeKnowledgeDocRepository) FindByID(ctx context.Context, id string) (models.KnowledgeDocument, error) {
	if f.err != nil {
		return models.KnowledgeDocument{}, f.err
	}
	for _, d := range f.docs {
		if d.ID == id {
			return d, nil
		}
	}
	return models.KnowledgeDocument{}, errors.New(constants.ErrKnowledgeDocNotFound)
}

func (f *fakeKnowledgeDocRepository) FindBySlug(ctx context.Context, slug string) (models.KnowledgeDocument, error) {
	if f.err != nil {
		return models.KnowledgeDocument{}, f.err
	}
	for _, d := range f.docs {
		if d.Slug == slug {
			return d, nil
		}
	}
	return models.KnowledgeDocument{}, errors.New(constants.ErrKnowledgeDocNotFound)
}

func (f *fakeKnowledgeDocRepository) Create(ctx context.Context, doc *models.KnowledgeDocument) error {
	if f.err != nil {
		return f.err
	}
	for _, d := range f.docs {
		if d.Slug == doc.Slug {
			return errors.New(constants.ErrKnowledgeDocSlugAlreadyExists)
		}
	}
	f.docs = append(f.docs, *doc)
	return nil
}

func (f *fakeKnowledgeDocRepository) Update(ctx context.Context, id string, doc *models.KnowledgeDocument) error {
	if f.err != nil {
		return f.err
	}
	for i, d := range f.docs {
		if d.ID == id {
			f.docs[i] = *doc
			return nil
		}
	}
	return errors.New(constants.ErrKnowledgeDocNotFound)
}

func (f *fakeKnowledgeDocRepository) Delete(ctx context.Context, id string) error {
	if f.err != nil {
		return f.err
	}
	for i, d := range f.docs {
		if d.ID == id {
			f.docs = append(f.docs[:i], f.docs[i+1:]...)
			return nil
		}
	}
	return errors.New(constants.ErrKnowledgeDocNotFound)
}

func (f *fakeKnowledgeDocRepository) Count(ctx context.Context) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return int64(len(f.docs)), nil
}

func TestKnowledgeDocumentService_GetDocuments(t *testing.T) {
	ctx := context.Background()
	repo := &fakeKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "Underwriting Doc", Category: models.KnowledgeCategoryUnderwriting, Status: models.IndexingStatusIndexed},
			{ID: "doc-2", Title: "Product Doc", Category: models.KnowledgeCategoryProduct, Status: models.IndexingStatusDraft},
		},
	}
	service := NewKnowledgeDocumentService(repo)

	// All docs
	all, err := service.GetDocuments(ctx, "all", "all", "")
	if err != nil {
		t.Fatalf("GetDocuments(all) error = %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("GetDocuments(all) len = %d, want 2", len(all))
	}

	// Filter category
	uwDocs, err := service.GetDocuments(ctx, "underwriting", "", "")
	if err != nil {
		t.Fatalf("GetDocuments(underwriting) error = %v", err)
	}
	if len(uwDocs) != 1 || uwDocs[0].ID != "doc-1" {
		t.Fatalf("GetDocuments(underwriting) = %+v, want doc-1", uwDocs)
	}

	// Invalid category query
	_, err = service.GetDocuments(ctx, "invalid-category", "", "")
	if err == nil {
		t.Fatal("expected error on invalid category, got nil")
	}
}

func TestKnowledgeDocumentService_GetByIDAndSlug(t *testing.T) {
	ctx := context.Background()
	repo := &fakeKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "Doc One", Slug: "doc-one"},
		},
	}
	service := NewKnowledgeDocumentService(repo)

	// GetByID success
	doc, err := service.GetDocumentByID(ctx, "doc-1")
	if err != nil {
		t.Fatalf("GetDocumentByID(doc-1) error = %v", err)
	}
	if doc.Title != "Doc One" {
		t.Fatalf("doc.Title = %q, want 'Doc One'", doc.Title)
	}

	// GetByID empty
	_, err = service.GetDocumentByID(ctx, "   ")
	if err == nil || err.Error() != constants.ErrKnowledgeDocIDRequired {
		t.Fatalf("expected ErrKnowledgeDocIDRequired, got %v", err)
	}

	// GetByID not found
	_, err = service.GetDocumentByID(ctx, "doc-99")
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("expected ErrKnowledgeDocNotFound, got %v", err)
	}

	// GetBySlug success
	docSlug, err := service.GetDocumentBySlug(ctx, "doc-one")
	if err != nil {
		t.Fatalf("GetDocumentBySlug(doc-one) error = %v", err)
	}
	if docSlug.ID != "doc-1" {
		t.Fatalf("docSlug.ID = %q, want 'doc-1'", docSlug.ID)
	}

	// GetBySlug empty
	_, err = service.GetDocumentBySlug(ctx, "")
	if err == nil || err.Error() != constants.ErrKnowledgeDocSlugInvalid {
		t.Fatalf("expected ErrKnowledgeDocSlugInvalid, got %v", err)
	}
}

func TestKnowledgeDocumentService_CreateDocument(t *testing.T) {
	ctx := context.Background()
	repo := &fakeKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-existing", Slug: "existing-slug"},
		},
	}
	service := NewKnowledgeDocumentService(repo)

	// 1. Validation error
	_, err := service.CreateDocument(ctx, dtos.CreateKnowledgeDocumentRequest{
		Title: "ab",
	})
	if err == nil {
		t.Fatal("expected validation error on short title, got nil")
	}

	// 2. Duplicate slug error
	_, err = service.CreateDocument(ctx, dtos.CreateKnowledgeDocumentRequest{
		Title:    "Existing Document",
		Slug:     "existing-slug",
		Category: "underwriting",
		Summary:  "Ringkasan dokumen yang sudah ada",
		Content:  "Konten lengkap dokumen yang sudah ada",
	})
	if err == nil || err.Error() != constants.ErrKnowledgeDocSlugAlreadyExists {
		t.Fatalf("expected ErrKnowledgeDocSlugAlreadyExists, got %v", err)
	}

	// 3. Success with auto slug generation & tags
	created, err := service.CreateDocument(ctx, dtos.CreateKnowledgeDocumentRequest{
		Title:    "SOP Verifikasi Dukcapil & Biometrik",
		Category: "compliance",
		Summary:  "Ringkasan verifikasi kependudukan",
		Content:  "Konten verifikasi kependudukan panjang...",
		Tags:     []string{"Dukcapil", "Compliance", ""},
	})
	if err != nil {
		t.Fatalf("CreateDocument error = %v", err)
	}
	if created.Slug != "sop-verifikasi-dukcapil-biometrik" {
		t.Fatalf("created.Slug = %q, want 'sop-verifikasi-dukcapil-biometrik'", created.Slug)
	}
	if created.ChunkCount < 2 {
		t.Fatalf("created.ChunkCount = %d, want at least 2", created.ChunkCount)
	}
	if created.Status != models.IndexingStatusIndexed {
		t.Fatalf("created.Status = %q, want 'indexed'", created.Status)
	}
}

func TestKnowledgeDocumentService_UpdateDocument(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	repo := &fakeKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "Original Title", Slug: "doc-one", Category: models.KnowledgeCategoryUnderwriting, Summary: "Original summary", Content: "Original content", CreatedAt: now, UpdatedAt: now},
			{ID: "doc-2", Title: "Second Doc", Slug: "doc-two", Category: models.KnowledgeCategoryProduct},
		},
	}
	service := NewKnowledgeDocumentService(repo)

	// 1. Empty ID
	_, err := service.UpdateDocument(ctx, "", dtos.UpdateKnowledgeDocumentRequest{})
	if err == nil || err.Error() != constants.ErrKnowledgeDocIDRequired {
		t.Fatalf("expected ErrKnowledgeDocIDRequired, got %v", err)
	}

	// 2. Not found
	_, err = service.UpdateDocument(ctx, "non-existent", dtos.UpdateKnowledgeDocumentRequest{})
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("expected ErrKnowledgeDocNotFound, got %v", err)
	}

	// 3. Slug conflict
	conflictSlug := "doc-two"
	_, err = service.UpdateDocument(ctx, "doc-1", dtos.UpdateKnowledgeDocumentRequest{
		Slug: &conflictSlug,
	})
	if err == nil || err.Error() != constants.ErrKnowledgeDocSlugAlreadyExists {
		t.Fatalf("expected ErrKnowledgeDocSlugAlreadyExists, got %v", err)
	}

	// 4. Successful update
	newTitle := "Updated Title"
	newContent := "Updated content with more text to test chunk count update..."
	updated, err := service.UpdateDocument(ctx, "doc-1", dtos.UpdateKnowledgeDocumentRequest{
		Title:   &newTitle,
		Content: &newContent,
	})
	if err != nil {
		t.Fatalf("UpdateDocument error = %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Fatalf("updated.Title = %q, want 'Updated Title'", updated.Title)
	}
	if updated.Content != newContent {
		t.Fatalf("updated.Content = %q, want %q", updated.Content, newContent)
	}
}

func TestKnowledgeDocumentService_DeleteDocument(t *testing.T) {
	ctx := context.Background()
	repo := &fakeKnowledgeDocRepository{
		docs: []models.KnowledgeDocument{
			{ID: "doc-1", Title: "Doc to Delete"},
		},
	}
	service := NewKnowledgeDocumentService(repo)

	// Empty ID
	err := service.DeleteDocument(ctx, "   ")
	if err == nil || err.Error() != constants.ErrKnowledgeDocIDRequired {
		t.Fatalf("expected ErrKnowledgeDocIDRequired, got %v", err)
	}

	// Success
	if err := service.DeleteDocument(ctx, "doc-1"); err != nil {
		t.Fatalf("DeleteDocument(doc-1) error = %v", err)
	}

	// Second delete fails with not found
	err = service.DeleteDocument(ctx, "doc-1")
	if err == nil || err.Error() != constants.ErrKnowledgeDocNotFound {
		t.Fatalf("expected ErrKnowledgeDocNotFound, got %v", err)
	}
}
