package dtos

import "time"

type ReindexDocumentResponse struct {
	DocumentID   string    `json:"document_id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	ChunkCount   int       `json:"chunk_count"`
	LastSyncedAt time.Time `json:"last_synced_at"`
}

type KnowledgeBaseMetricsResponse struct {
	TotalDocuments      int64            `json:"total_documents"`
	TotalChunks         int64            `json:"total_chunks"`
	IndexedDocuments    int64            `json:"indexed_documents"`
	SyncingDocuments    int64            `json:"syncing_documents"`
	DraftDocuments      int64            `json:"draft_documents"`
	AverageChunksPerDoc float64          `json:"average_chunks_per_doc"`
	CategoryBreakdown   map[string]int64 `json:"category_breakdown"`
	LastSyncTime        *time.Time       `json:"last_sync_time"`
}

type SimulateChatRequest struct {
	Query    string `json:"query"`
	Category string `json:"category,omitempty"`
	TopK     int    `json:"top_k,omitempty"`
}

type RetrievedChunkDetail struct {
	ID         string  `json:"id"`
	DocumentID string  `json:"document_id"`
	Title      string  `json:"title"`
	SourceType string  `json:"source_type"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
	ChunkIndex int     `json:"chunk_index"`
}

type SimulateChatResponse struct {
	Query           string                 `json:"query"`
	Answer          string                 `json:"answer"`
	RetrievedChunks []RetrievedChunkDetail `json:"retrieved_chunks"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
}
