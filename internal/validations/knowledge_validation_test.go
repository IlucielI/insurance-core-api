package validations

import (
	"reflect"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
)

func TestValidateKnowledgeCategory(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		expectErr bool
	}{
		{"Valid underwriting", "underwriting", false},
		{"Valid product", "product", false},
		{"Valid claim_faq", "claim_faq", false},
		{"Valid compliance", "compliance", false},
		{"Valid company", "company", false},
		{"Invalid empty", "", true},
		{"Invalid random", "non_existent_category", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKnowledgeCategory(tt.category)
			if tt.expectErr && err == nil {
				t.Fatalf("expected error for category %q, got nil", tt.category)
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error for category %q: %v", tt.category, err)
			}
		})
	}
}

func TestValidateIndexingStatus(t *testing.T) {
	if err := ValidateIndexingStatus("indexed"); err != nil {
		t.Fatalf("expected valid indexed, got %v", err)
	}
	if err := ValidateIndexingStatus("syncing"); err != nil {
		t.Fatalf("expected valid syncing, got %v", err)
	}
	if err := ValidateIndexingStatus("draft"); err != nil {
		t.Fatalf("expected valid draft, got %v", err)
	}
	if err := ValidateIndexingStatus("invalid"); err == nil {
		t.Fatalf("expected error for invalid status, got nil")
	}
	if err := ValidateIndexingStatus(""); err == nil {
		t.Fatalf("expected error for empty status, got nil")
	}
}

func TestValidateCreateKnowledgeDocRequest(t *testing.T) {
	t.Run("Nil request", func(t *testing.T) {
		err := ValidateCreateKnowledgeDocRequest(nil)
		if err == nil || err.Error() != constants.ErrKnowledgeDocRequestBodyInvalid {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocRequestBodyInvalid, err)
		}
	})

	t.Run("Missing title", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Category: "underwriting",
			Summary:  "summary text",
			Content:  "content text",
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocTitleRequired {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocTitleRequired, err)
		}
	})

	t.Run("Title too short", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Title:    "ab",
			Category: "underwriting",
			Summary:  "summary text",
			Content:  "content text",
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocTitleInvalid {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocTitleInvalid, err)
		}
	})

	t.Run("Invalid category", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Title:    "Valid Title",
			Category: "unknown_category",
			Summary:  "summary text",
			Content:  "content text",
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocCategoryInvalid {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocCategoryInvalid, err)
		}
	})

	t.Run("Missing summary", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Title:    "Valid Title",
			Category: "underwriting",
			Content:  "content text",
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocSummaryRequired {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocSummaryRequired, err)
		}
	})

	t.Run("Missing content", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Title:    "Valid Title",
			Category: "underwriting",
			Summary:  "summary text",
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocContentRequired {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocContentRequired, err)
		}
	})

	t.Run("Invalid slug format", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Title:    "Valid Title",
			Slug:     "Invalid Slug!",
			Category: "underwriting",
			Summary:  "summary text",
			Content:  "content text",
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocSlugInvalid {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocSlugInvalid, err)
		}
	})

	t.Run("Valid request with tags and defaults", func(t *testing.T) {
		req := &dtos.CreateKnowledgeDocumentRequest{
			Title:    "Valid Title",
			Slug:     "valid-slug",
			Category: "underwriting",
			Summary:  "summary text",
			Content:  "content text",
			Tags:     []string{"Underwriting", " SOP ", ""},
		}
		err := ValidateCreateKnowledgeDocRequest(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if req.Slug != "valid-slug" {
			t.Fatalf("expected slug 'valid-slug', got %q", req.Slug)
		}
		if req.Status != "indexed" {
			t.Fatalf("expected status 'indexed', got %q", req.Status)
		}
		expectedTags := []string{"underwriting", "sop"}
		if !reflect.DeepEqual(req.Tags, expectedTags) {
			t.Fatalf("expected tags %v, got %v", expectedTags, req.Tags)
		}
	})
}

func TestValidateUpdateKnowledgeDocRequest(t *testing.T) {
	t.Run("Nil request", func(t *testing.T) {
		err := ValidateUpdateKnowledgeDocRequest(nil)
		if err == nil || err.Error() != constants.ErrKnowledgeDocRequestBodyInvalid {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocRequestBodyInvalid, err)
		}
	})

	t.Run("Valid partial update", func(t *testing.T) {
		newTitle := "Updated Title"
		newSlug := "updated-slug"
		req := &dtos.UpdateKnowledgeDocumentRequest{
			Title: &newTitle,
			Slug:  &newSlug,
		}
		err := ValidateUpdateKnowledgeDocRequest(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if *req.Title != "Updated Title" {
			t.Fatalf("expected title 'Updated Title', got %q", *req.Title)
		}
		if *req.Slug != "updated-slug" {
			t.Fatalf("expected slug 'updated-slug', got %q", *req.Slug)
		}
	})

	t.Run("Empty title in update", func(t *testing.T) {
		emptyTitle := "  "
		req := &dtos.UpdateKnowledgeDocumentRequest{
			Title: &emptyTitle,
		}
		err := ValidateUpdateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocTitleRequired {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocTitleRequired, err)
		}
	})

	t.Run("Invalid slug in update", func(t *testing.T) {
		badSlug := "bad slug!"
		req := &dtos.UpdateKnowledgeDocumentRequest{
			Slug: &badSlug,
		}
		err := ValidateUpdateKnowledgeDocRequest(req)
		if err == nil || err.Error() != constants.ErrKnowledgeDocSlugInvalid {
			t.Fatalf("expected %q, got %v", constants.ErrKnowledgeDocSlugInvalid, err)
		}
	})
}

func TestValidateKnowledgeListQuery(t *testing.T) {
	cat, st, err := ValidateKnowledgeListQuery("underwriting", "indexed")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cat != "underwriting" || st != "indexed" {
		t.Fatalf("expected 'underwriting' and 'indexed', got %q and %q", cat, st)
	}

	catAll, stAll, errAll := ValidateKnowledgeListQuery("all", "all")
	if errAll != nil {
		t.Fatalf("expected no error, got %v", errAll)
	}
	if catAll != "" || stAll != "" {
		t.Fatalf("expected empty strings for 'all', got %q and %q", catAll, stAll)
	}

	_, _, errInvalidCat := ValidateKnowledgeListQuery("bad_cat", "")
	if errInvalidCat == nil {
		t.Fatalf("expected error for invalid category, got nil")
	}

	_, _, errInvalidSt := ValidateKnowledgeListQuery("", "bad_status")
	if errInvalidSt == nil {
		t.Fatalf("expected error for invalid status, got nil")
	}
}
