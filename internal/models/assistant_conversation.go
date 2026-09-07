package models

import (
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
)

type AssistantConversation struct {
	ID        string             `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Title     string             `gorm:"type:varchar(255);not null;default:''" json:"title"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Messages  []AssistantMessage `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

func (AssistantConversation) TableName() string { return "assistant_conversations" }

type AssistantMessage struct {
	ID             string         `gorm:"primaryKey;type:varchar(64)" json:"id"`
	ConversationID string         `gorm:"type:varchar(64);not null;index:idx_assistant_messages_conversation" json:"conversation_id"`
	Role           string         `gorm:"type:varchar(32);not null" json:"role"`
	Content        string         `gorm:"type:text" json:"content"`
	ToolCalls      []llm.ToolCall `gorm:"type:jsonb;serializer:json" json:"tool_calls,omitempty"`
	ToolCallID     string         `gorm:"type:varchar(64)" json:"tool_call_id,omitempty"`
	CreatedAt      time.Time      `gorm:"index:idx_assistant_messages_conversation" json:"created_at"`
}

func (AssistantMessage) TableName() string { return "assistant_messages" }
