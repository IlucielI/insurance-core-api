package dtos

import (
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

type CreateKnowledgeDocumentRequest struct {
	Title    string   `json:"title"`
	Slug     string   `json:"slug"`
	Category string   `json:"category"`
	Summary  string   `json:"summary"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Status   string   `json:"status,omitempty"`
}

type UpdateKnowledgeDocumentRequest struct {
	Title    *string   `json:"title,omitempty"`
	Slug     *string   `json:"slug,omitempty"`
	Category *string   `json:"category,omitempty"`
	Summary  *string   `json:"summary,omitempty"`
	Content  *string   `json:"content,omitempty"`
	Tags     *[]string `json:"tags,omitempty"`
	Status   *string   `json:"status,omitempty"`
}

type KnowledgeDocumentResponse struct {
	ID           string                   `json:"id"`
	Title        string                   `json:"title"`
	Slug         string                   `json:"slug"`
	Category     models.KnowledgeCategory `json:"category"`
	Summary      string                   `json:"summary"`
	Content      string                   `json:"content"`
	Tags         []string                 `json:"tags"`
	ChunkCount   int                      `json:"chunk_count"`
	Status       models.IndexingStatus    `json:"status"`
	LastSyncedAt *time.Time               `json:"last_synced_at"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

func ToKnowledgeDocumentResponse(doc models.KnowledgeDocument) KnowledgeDocumentResponse {
	tags := doc.Tags
	if tags == nil {
		tags = []string{}
	}
	return KnowledgeDocumentResponse{
		ID:           doc.ID,
		Title:        doc.Title,
		Slug:         doc.Slug,
		Category:     doc.Category,
		Summary:      doc.Summary,
		Content:      doc.Content,
		Tags:         tags,
		ChunkCount:   doc.ChunkCount,
		Status:       doc.Status,
		LastSyncedAt: doc.LastSyncedAt,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}

func ToKnowledgeDocumentResponses(docs []models.KnowledgeDocument) []KnowledgeDocumentResponse {
	res := make([]KnowledgeDocumentResponse, len(docs))
	for i, doc := range docs {
		res[i] = ToKnowledgeDocumentResponse(doc)
	}
	return res
}
