package models

import "time"

type KnowledgeCategory string

const (
	KnowledgeCategoryUnderwriting KnowledgeCategory = "underwriting"
	KnowledgeCategoryProduct      KnowledgeCategory = "product"
	KnowledgeCategoryClaimFAQ     KnowledgeCategory = "claim_faq"
	KnowledgeCategoryCompliance   KnowledgeCategory = "compliance"
	KnowledgeCategoryCompany      KnowledgeCategory = "company"
)

type IndexingStatus string

const (
	IndexingStatusIndexed IndexingStatus = "indexed"
	IndexingStatusSyncing IndexingStatus = "syncing"
	IndexingStatusDraft   IndexingStatus = "draft"
)

type KnowledgeDocument struct {
	ID           string            `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Title        string            `gorm:"type:varchar(255);not null" json:"title"`
	Slug         string            `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Category     KnowledgeCategory `gorm:"type:varchar(64);not null;index" json:"category"`
	Summary      string            `gorm:"type:text;not null" json:"summary"`
	Content      string            `gorm:"type:text;not null" json:"content"`
	Tags         []string          `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"tags"`
	ChunkCount   int               `gorm:"not null;default:0" json:"chunk_count"`
	Status       IndexingStatus    `gorm:"type:varchar(32);not null;default:'indexed';index" json:"status"`
	LastSyncedAt *time.Time        `json:"last_synced_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

func (KnowledgeDocument) TableName() string {
	return "knowledge_documents"
}
