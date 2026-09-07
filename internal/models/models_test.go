package models

import "testing"

func TestTableNames(t *testing.T) {
	product := Product{}
	if product.TableName() != "products" {
		t.Fatalf("Product.TableName() = %q, want products", product.TableName())
	}

	application := Application{}
	if application.TableName() != "applications" {
		t.Fatalf("Application.TableName() = %q, want applications", application.TableName())
	}

	reviewCheck := ApplicationReviewCheck{}
	if reviewCheck.TableName() != "application_review_checks" {
		t.Fatalf("ApplicationReviewCheck.TableName() = %q, want application_review_checks", reviewCheck.TableName())
	}
}

func TestKnowledgeChunkTableName(t *testing.T) {
	chunk := KnowledgeChunk{}
	if chunk.TableName() != "knowledge_chunks" {
		t.Fatalf("KnowledgeChunk.TableName() = %q, want knowledge_chunks", chunk.TableName())
	}
}

func TestAssistantConversationTableNames(t *testing.T) {
	conv := AssistantConversation{}
	if conv.TableName() != "assistant_conversations" {
		t.Fatalf("AssistantConversation.TableName() = %q, want assistant_conversations", conv.TableName())
	}

	msg := AssistantMessage{}
	if msg.TableName() != "assistant_messages" {
		t.Fatalf("AssistantMessage.TableName() = %q, want assistant_messages", msg.TableName())
	}
}

func TestPricingRulesSerialization(t *testing.T) {
	rules := PricingRules{
		BaseRate:           0.0035,
		SumAssuredPresets:  []int64{100_000_000, 250_000_000, 500_000_000, 1_000_000_000},
		PaymentTermPresets: []int{5, 10, 15, 20},
	}
	if len(rules.SumAssuredPresets) != 4 {
		t.Fatalf("SumAssuredPresets len = %d, want 4", len(rules.SumAssuredPresets))
	}
	if len(rules.PaymentTermPresets) != 4 {
		t.Fatalf("PaymentTermPresets len = %d, want 4", len(rules.PaymentTermPresets))
	}
}
