package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
)

func TestPostgresAssistantConversationRepository(t *testing.T) {
	db := sqliteDB(t)
	if err := db.AutoMigrate(&models.AssistantConversation{}, &models.AssistantMessage{}); err != nil {
		t.Fatalf("AutoMigrate() error = %v", err)
	}

	cache := newFakeCache()
	repo := NewPostgresAssistantConversationRepository(db, cache)

	ctx := context.Background()

	// 1. GetOrCreateConversation
	conv, err := repo.GetOrCreateConversation(ctx, "conv-1", "Insurance Discussion")
	if err != nil {
		t.Fatalf("GetOrCreateConversation() error = %v", err)
	}
	if conv.ID != "conv-1" || conv.Title != "Insurance Discussion" {
		t.Fatalf("unexpected conv: %+v", conv)
	}

	// 2. SaveMessage
	msg1 := models.AssistantMessage{
		ID:             "msg-1",
		ConversationID: "conv-1",
		Role:           "user",
		Content:        "Hello, I want insurance.",
		CreatedAt:      time.Now().Add(-2 * time.Minute),
	}
	if err := repo.SaveMessage(ctx, msg1); err != nil {
		t.Fatalf("SaveMessage() error = %v", err)
	}

	// 3. SaveMessages (batch)
	msg2 := models.AssistantMessage{
		ID:             "msg-2",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "Sure, what product do you need?",
		CreatedAt:      time.Now().Add(-1 * time.Minute),
	}
	msg3 := models.AssistantMessage{
		ID:             "msg-3",
		ConversationID: "conv-1",
		Role:           "assistant",
		Content:        "",
		ToolCalls: []llm.ToolCall{
			{
				ID:   "call-1",
				Type: "function",
				Function: llm.ToolCallFunction{
					Name:      "list_products",
					Arguments: `{}`,
				},
			},
		},
		CreatedAt: time.Now(),
	}
	if err := repo.SaveMessages(ctx, []models.AssistantMessage{msg2, msg3}); err != nil {
		t.Fatalf("SaveMessages() error = %v", err)
	}

	// 4. ListMessages
	messages, err := repo.ListMessages(ctx, "conv-1", 0)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 3 {
		t.Fatalf("len(messages) = %d, want 3", len(messages))
	}

	// ListMessages with limit (most recent)
	recent, err := repo.ListMessages(ctx, "conv-1", 2)
	if err != nil {
		t.Fatalf("ListMessages(limit=2) error = %v", err)
	}
	if len(recent) != 2 || recent[0].ID != "msg-2" || recent[1].ID != "msg-3" {
		t.Fatalf("unexpected recent messages: %+v", recent)
	}

	// 5. GetConversation with preloaded messages
	fetchedConv, err := repo.GetConversation(ctx, "conv-1")
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if len(fetchedConv.Messages) != 3 {
		t.Fatalf("len(fetchedConv.Messages) = %d, want 3", len(fetchedConv.Messages))
	}
	if fetchedConv.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}

	// 6. DeleteConversation
	if err := repo.DeleteConversation(ctx, "conv-1"); err != nil {
		t.Fatalf("DeleteConversation() error = %v", err)
	}

	_, err = repo.GetConversation(ctx, "conv-1")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Fatalf("GetConversation() after delete error = %v, want ErrConversationNotFound", err)
	}
}
