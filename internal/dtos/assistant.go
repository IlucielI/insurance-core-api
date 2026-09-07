package dtos

import "github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"

type AssistantChatRequest struct {
	ConversationID string               `json:"conversation_id,omitempty"`
	Message        string               `json:"message"`
	ProductSlug    string               `json:"product_slug,omitempty"`
	Quote          *ProductQuoteRequest `json:"quote,omitempty"`
	Slug           string               `json:"slug,omitempty"`
}

type AssistantSource struct {
	Title      string  `json:"title"`
	SourceType string  `json:"source_type"`
	Score      float64 `json:"score"`
	Excerpt    string  `json:"excerpt"`
}

type AssistantChatResponse struct {
	ConversationID string            `json:"conversation_id"`
	Answer         string            `json:"answer"`
	Sources        []AssistantSource `json:"sources"`
	ToolsUsed      []string          `json:"tools_used,omitempty"`
}

type AssistantMessageResponse struct {
	ID         string         `json:"id"`
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []llm.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	CreatedAt  string         `json:"created_at"`
}

type AssistantConversationResponse struct {
	ID        string                     `json:"id"`
	Title     string                     `json:"title"`
	CreatedAt string                     `json:"created_at"`
	UpdatedAt string                     `json:"updated_at"`
	Messages  []AssistantMessageResponse `json:"messages"`
}
