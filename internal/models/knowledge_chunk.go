package models

import (
	"time"

	pgvector "github.com/pgvector/pgvector-go"
)

type KnowledgeChunk struct {
	ID         string          `gorm:"primaryKey;type:varchar(64)" json:"id"`
	SourceType string          `gorm:"type:varchar(64);not null;index" json:"source_type"`
	Title      string          `gorm:"type:varchar(120);not null" json:"title"`
	Content    string          `gorm:"type:text;not null" json:"content"`
	ChunkIndex int             `gorm:"not null" json:"chunk_index"`
	Embedding  pgvector.Vector `gorm:"type:vector(1024)" json:"-"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func (KnowledgeChunk) TableName() string { return "knowledge_chunks" }
