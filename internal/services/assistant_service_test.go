package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
)

type assistantLLMFake struct {
	answer                string
	embedding             []float32
	embedErr, chatErr     error
	embedCalls, chatCalls int
}

func (f *assistantLLMFake) CreateEmbedding(context.Context, llm.EmbeddingInput) ([]float32, error) {
	f.embedCalls++
	return f.embedding, f.embedErr
}
func (f *assistantLLMFake) CreateChatCompletion(context.Context, llm.ChatCompletionInput) (string, error) {
	f.chatCalls++
	return f.answer, f.chatErr
}

type knowledgeFake struct {
	matches                         []repositories.KnowledgeChunkMatch
	count                           int64
	searched                        bool
	replaced                        []models.KnowledgeChunk
	countErr, searchErr, replaceErr error
}

func (f *knowledgeFake) Count(context.Context) (int64, error) { return f.count, f.countErr }
func (f *knowledgeFake) Search(context.Context, []float32, int) ([]repositories.KnowledgeChunkMatch, error) {
	f.searched = true
	return f.matches, f.searchErr
}
func (f *knowledgeFake) ReplaceAll(_ context.Context, c []models.KnowledgeChunk) error {
	f.replaced = c
	return f.replaceErr
}

func TestAssistantChatFiltersDistantSources(t *testing.T) {
	repo := &knowledgeFake{matches: []repositories.KnowledgeChunkMatch{{KnowledgeChunk: models.KnowledgeChunk{Title: "relevant", Content: "answer", SourceType: "faq"}, Distance: .2}, {KnowledgeChunk: models.KnowledgeChunk{Title: "noise", Content: "noise"}, Distance: .9}}}
	model := &assistantLLMFake{answer: "ok", embedding: embeddingFixture()}
	got, err := NewAssistantService(repo, model, nil).Chat(context.Background(), " question ")
	if err != nil || got.Answer != "ok" || len(got.Sources) != 1 || !repo.searched || model.chatCalls != 1 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}
func TestAssistantChatRejectsInvalidInput(t *testing.T) {
	_, err := NewAssistantService(nil, nil, nil).Chat(context.Background(), " ")
	if !errors.Is(err, constants.ErrAssistantMessageRequiredError) {
		t.Fatal("expected validation error")
	}
}
func TestSeedDefaultKnowledgeSkipsExisting(t *testing.T) {
	repo := &knowledgeFake{count: 1}
	model := &assistantLLMFake{embedding: embeddingFixture()}
	if err := NewAssistantService(repo, model, nil).SeedDefaultKnowledge(context.Background()); err != nil {
		t.Fatal(err)
	}
	if model.embedCalls != 0 || repo.replaced != nil {
		t.Fatal("existing knowledge was reseeded")
	}
}

func TestAssistantErrorPaths(t *testing.T) {
	ctx := context.Background()
	if err := NewAssistantService(nil, nil, nil).SeedDefaultKnowledge(ctx); !errors.Is(err, constants.ErrAssistantServiceUnavailableError) {
		t.Fatal(err)
	}
	repo := &knowledgeFake{countErr: errors.New("count")}
	model := &assistantLLMFake{embedding: embeddingFixture()}
	if err := NewAssistantService(repo, model, nil).SeedDefaultKnowledge(ctx); err == nil {
		t.Fatal("expected count error")
	}
	repo.countErr = nil
	model.embedding = nil
	if err := NewAssistantService(repo, model, nil).SeedDefaultKnowledge(ctx); err == nil {
		t.Fatal("expected empty embedding error")
	}
	model.embedding = embeddingFixture()
	model.embedErr = errors.New("embed")
	if err := NewAssistantService(repo, model, nil).SeedDefaultKnowledge(ctx); err == nil {
		t.Fatal("expected embed error")
	}
	repo2 := &knowledgeFake{searchErr: errors.New("search")}
	model2 := &assistantLLMFake{embedding: embeddingFixture()}
	if _, err := NewAssistantService(repo2, model2, nil).Chat(ctx, "hello"); err == nil {
		t.Fatal("expected search error")
	}
}

func TestAssistantChatReturnsCompletionError(t *testing.T) {
	repo := &knowledgeFake{}
	model := &assistantLLMFake{embedding: embeddingFixture(), chatErr: errors.New("completion")}
	if _, err := NewAssistantService(repo, model, nil).Chat(context.Background(), "hello"); err == nil {
		t.Fatal("expected completion error")
	}
}

func TestAssistantSeedReturnsRepositoryError(t *testing.T) {
	repo := &knowledgeFake{replaceErr: errors.New("replace")}
	model := &assistantLLMFake{embedding: embeddingFixture()}
	if err := NewAssistantService(repo, model, nil).SeedDefaultKnowledge(context.Background()); err == nil {
		t.Fatal("expected replace error")
	}
}

func TestAssistantKnowledgeHelpers(t *testing.T) {
	if len(defaultKnowledgeChunks()) == 0 || len(chunkKnowledge(defaultKnowledgeChunks())) == 0 {
		t.Fatal("default knowledge is empty")
	}
	if got := excerpt(strings.Repeat("x", 181)); len(got) != 183 {
		t.Fatalf("excerpt length = %d, want 183", len(got))
	}
	if max(1, 2) != 2 {
		t.Fatal("max returned wrong value")
	}
}

func TestAssistantChatIncludesQuoteContext(t *testing.T) {
	repo := &knowledgeFake{}
	model := &assistantLLMFake{answer: "quote answer", embedding: embeddingFixture()}
	products := &fakeProductRepository{product: productFixture()}
	service := NewAssistantService(repo, model, NewProductService(products))
	request := dtos.ProductQuoteRequest(quoteInputFixture())

	got, err := service.ChatWithQuote(context.Background(), "hitung premi", &request, "secure-life-plus")
	if err != nil {
		t.Fatalf("ChatWithQuote() error = %v", err)
	}
	if got.Answer != "quote answer" || model.chatCalls != 1 {
		t.Fatalf("ChatWithQuote() = %+v, chatCalls=%d", got, model.chatCalls)
	}
}

func TestAssistantChatWithQuoteRequiresQuoteServiceAndSlug(t *testing.T) {
	repo := &knowledgeFake{}
	model := &assistantLLMFake{answer: "ok", embedding: embeddingFixture()}
	request := dtos.ProductQuoteRequest(quoteInputFixture())

	_, err := NewAssistantService(repo, model, nil).ChatWithQuote(context.Background(), "hitung premi", &request, "secure-life-plus")
	if !errors.Is(err, constants.ErrAssistantServiceUnavailableError) {
		t.Fatalf("ChatWithQuote() error = %v, want unavailable", err)
	}

	_, err = NewAssistantService(repo, model, NewProductService(&fakeProductRepository{product: productFixture()})).ChatWithQuote(context.Background(), "hitung premi", &request, " ")
	if err == nil {
		t.Fatal("expected slug error")
	}
}

func embeddingFixture() []float32 {
	return make([]float32, constants.AssistantEmbeddingDimension)
}
