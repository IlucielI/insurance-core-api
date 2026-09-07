package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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
	toolCalls             []llm.ToolCall
	chatWithToolsFn       func(context.Context, llm.ChatCompletionInput) (llm.ChatCompletionOutput, error)
	streamTokens          []string
	streamErr             error
	streamCalls           int
}

func (f *assistantLLMFake) StreamChatCompletion(ctx context.Context, in llm.ChatCompletionInput, onToken func(string) error) error {
	f.streamCalls++
	if f.streamErr != nil {
		return f.streamErr
	}
	if len(f.streamTokens) > 0 {
		for _, t := range f.streamTokens {
			if err := onToken(t); err != nil {
				return err
			}
		}
		return nil
	}
	if f.answer != "" {
		return onToken(f.answer)
	}
	return nil
}

func (f *assistantLLMFake) CreateEmbedding(context.Context, llm.EmbeddingInput) ([]float32, error) {
	f.embedCalls++
	return f.embedding, f.embedErr
}
func (f *assistantLLMFake) CreateChatCompletion(context.Context, llm.ChatCompletionInput) (string, error) {
	f.chatCalls++
	return f.answer, f.chatErr
}
func (f *assistantLLMFake) CreateChatCompletionWithTools(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
	if f.chatWithToolsFn != nil {
		return f.chatWithToolsFn(ctx, in)
	}
	f.chatCalls++
	if f.chatErr != nil {
		return llm.ChatCompletionOutput{}, f.chatErr
	}
	if len(f.toolCalls) > 0 {
		tc := f.toolCalls
		f.toolCalls = nil
		return llm.ChatCompletionOutput{
			ToolCalls: tc,
		}, nil
	}
	return llm.ChatCompletionOutput{Content: f.answer}, nil
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

func TestAssistantChatWithTools_CalculateQuote(t *testing.T) {
	repo := &knowledgeFake{}
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	// Simulated two-turn tool conversation:
	// Turn 1: LLM returns tool_call calculate_quote
	// Turn 2: LLM receives tool output and returns final natural language summary
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			// Check if previous message has role "tool" (second iteration)
			if len(in.Messages) > 0 && in.Messages[len(in.Messages)-1].Role == "tool" {
				toolMsg := in.Messages[len(in.Messages)-1]
				if !strings.Contains(toolMsg.Content, "estimated_premium") {
					t.Fatalf("tool message does not contain estimated_premium: %s", toolMsg.Content)
				}
				return llm.ChatCompletionOutput{
					Content: "Estimasi premi Anda adalah Rp 100.000 per bulan.",
				}, nil
			}

			// First iteration: request calculate_quote
			return llm.ChatCompletionOutput{
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_quote_1",
						Type: "function",
						Function: llm.ToolCallFunction{
							Name:      "calculate_quote",
							Arguments: `{"product_slug":"secure-life-plus","age":30,"gender":"male","sum_assured":100000000,"payment_term":10,"payment_frequency":"monthly"}`,
						},
					},
				},
			}, nil
		},
	}

	service := NewAssistantService(repo, model, productService)
	resp, err := service.Chat(context.Background(), "Hitungkan premi asuransi secure life plus untuk saya")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if resp.Answer != "Estimasi premi Anda adalah Rp 100.000 per bulan." {
		t.Fatalf("Chat() answer = %q", resp.Answer)
	}
	if len(resp.ToolsUsed) != 1 || resp.ToolsUsed[0] != "calculate_quote" {
		t.Fatalf("ToolsUsed = %+v, want ['calculate_quote']", resp.ToolsUsed)
	}
}

func TestAssistantChatWithTools_ListProducts(t *testing.T) {
	repo := &knowledgeFake{}
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			if len(in.Messages) > 0 && in.Messages[len(in.Messages)-1].Role == "tool" {
				return llm.ChatCompletionOutput{
					Content: "Produk yang tersedia adalah Secure Life Plus.",
				}, nil
			}

			return llm.ChatCompletionOutput{
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_list_1",
						Type: "function",
						Function: llm.ToolCallFunction{
							Name:      "list_products",
							Arguments: `{"category":"life"}`,
						},
					},
				},
			}, nil
		},
	}

	service := NewAssistantService(repo, model, productService)
	resp, err := service.Chat(context.Background(), "Ada produk asuransi apa saja?")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if resp.Answer != "Produk yang tersedia adalah Secure Life Plus." {
		t.Fatalf("Chat() answer = %q", resp.Answer)
	}
	if len(resp.ToolsUsed) != 1 || resp.ToolsUsed[0] != "list_products" {
		t.Fatalf("ToolsUsed = %+v, want ['list_products']", resp.ToolsUsed)
	}
}

func TestAssistantChatWithTools_ToolErrorHandled(t *testing.T) {
	repo := &knowledgeFake{}
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			if len(in.Messages) > 0 && in.Messages[len(in.Messages)-1].Role == "tool" {
				toolMsg := in.Messages[len(in.Messages)-1]
				if !strings.Contains(toolMsg.Content, "error") {
					t.Fatalf("expected error in tool message: %s", toolMsg.Content)
				}
				return llm.ChatCompletionOutput{
					Content: "Mohon maaf, uang pertanggungan di luar batas.",
				}, nil
			}

			// Out of range sum_assured to trigger error in calculate_quote
			return llm.ChatCompletionOutput{
				ToolCalls: []llm.ToolCall{
					{
						ID:   "call_err_1",
						Type: "function",
						Function: llm.ToolCallFunction{
							Name:      "calculate_quote",
							Arguments: `{"product_slug":"secure-life-plus","age":30,"gender":"male","sum_assured":1,"payment_term":10}`,
						},
					},
				},
			}, nil
		},
	}

	service := NewAssistantService(repo, model, productService)
	resp, err := service.Chat(context.Background(), "Hitung premi 1 rupiah")
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if resp.Answer != "Mohon maaf, uang pertanggungan di luar batas." {
		t.Fatalf("Chat() answer = %q", resp.Answer)
	}
	if len(resp.ToolsUsed) != 1 || resp.ToolsUsed[0] != "calculate_quote" {
		t.Fatalf("ToolsUsed = %+v", resp.ToolsUsed)
	}
}

type fakeConversationRepository struct {
	conversations map[string]models.AssistantConversation
	messages      map[string][]models.AssistantMessage
}

func newFakeConversationRepository() *fakeConversationRepository {
	return &fakeConversationRepository{
		conversations: make(map[string]models.AssistantConversation),
		messages:      make(map[string][]models.AssistantMessage),
	}
}

func (f *fakeConversationRepository) GetOrCreateConversation(ctx context.Context, id string, title string) (models.AssistantConversation, error) {
	if conv, ok := f.conversations[id]; ok {
		return conv, nil
	}
	conv := models.AssistantConversation{ID: id, Title: title, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	f.conversations[id] = conv
	return conv, nil
}

func (f *fakeConversationRepository) GetConversation(ctx context.Context, id string) (models.AssistantConversation, error) {
	conv, ok := f.conversations[id]
	if !ok {
		return models.AssistantConversation{}, repositories.ErrConversationNotFound
	}
	conv.Messages = f.messages[id]
	return conv, nil
}

func (f *fakeConversationRepository) ListMessages(ctx context.Context, conversationID string, limit int) ([]models.AssistantMessage, error) {
	msgs := f.messages[conversationID]
	if limit > 0 && len(msgs) > limit {
		return msgs[len(msgs)-limit:], nil
	}
	return msgs, nil
}

func (f *fakeConversationRepository) SaveMessage(ctx context.Context, msg models.AssistantMessage) error {
	return f.SaveMessages(ctx, []models.AssistantMessage{msg})
}

func (f *fakeConversationRepository) SaveMessages(ctx context.Context, msgs []models.AssistantMessage) error {
	for _, msg := range msgs {
		f.messages[msg.ConversationID] = append(f.messages[msg.ConversationID], msg)
	}
	return nil
}

func (f *fakeConversationRepository) DeleteConversation(ctx context.Context, id string) error {
	delete(f.conversations, id)
	delete(f.messages, id)
	return nil
}

func TestAssistantChatWithConversationMultiTurn(t *testing.T) {
	repo := &knowledgeFake{}
	convRepo := newFakeConversationRepository()
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	var lastMessagesReceived []llm.Message
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			lastMessagesReceived = in.Messages
			return llm.ChatCompletionOutput{Content: "Saya siap membantu."}, nil
		},
	}

	service := NewAssistantService(repo, model, productService, convRepo)

	// Turn 1: Starts new conversation
	resp1, err := service.Chat(context.Background(), "Halo, saya tertarik produk Secure Life Plus.")
	if err != nil {
		t.Fatalf("Turn 1 error: %v", err)
	}
	if resp1.ConversationID == "" {
		t.Fatal("expected non-empty ConversationID")
	}

	// Turn 2: Uses conversation ID from Turn 1
	_, err = service.ChatWithConversation(context.Background(), "Berapa preminya?", resp1.ConversationID, nil, "")
	if err != nil {
		t.Fatalf("Turn 2 error: %v", err)
	}

	// Check that Turn 2 messages sent to LLM contains history from Turn 1
	var foundTurn1User bool
	for _, m := range lastMessagesReceived {
		if strings.Contains(m.Content, "Halo, saya tertarik produk Secure Life Plus.") {
			foundTurn1User = true
			break
		}
	}
	if !foundTurn1User {
		t.Fatalf("history from Turn 1 was not included in Turn 2 prompt: %+v", lastMessagesReceived)
	}

	// Test GetConversation
	conv, err := service.GetConversation(context.Background(), resp1.ConversationID)
	if err != nil {
		t.Fatalf("GetConversation error: %v", err)
	}
	if len(conv.Messages) < 4 {
		t.Fatalf("expected at least 4 messages in conversation, got %d", len(conv.Messages))
	}

	// Test DeleteConversation
	if err := service.DeleteConversation(context.Background(), resp1.ConversationID); err != nil {
		t.Fatalf("DeleteConversation error: %v", err)
	}
	_, err = service.GetConversation(context.Background(), resp1.ConversationID)
	if !errors.Is(err, repositories.ErrConversationNotFound) {
		t.Fatalf("expected ErrConversationNotFound, got %v", err)
	}
}

func TestAssistantChat_SanitizesLeadingOrphanToolMessages(t *testing.T) {
	repo := &knowledgeFake{}
	convRepo := newFakeConversationRepository()
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	// Pre-populate conversation with an orphan tool message at the start
	convID := "orphan-conv"
	if _, err := convRepo.GetOrCreateConversation(context.Background(), convID, "Test"); err != nil {
		t.Fatalf("setup GetOrCreateConversation error: %v", err)
	}
	if err := convRepo.SaveMessages(context.Background(), []models.AssistantMessage{
		{
			ID:             "orphan-tool-msg",
			ConversationID: convID,
			Role:           "tool",
			Content:        `{"quote":100000}`,
			ToolCallID:     "call_xyz",
			CreatedAt:      time.Now().Add(-10 * time.Minute),
		},
		{
			ID:             "subsequent-user-msg",
			ConversationID: convID,
			Role:           "user",
			Content:        "Apakah ada promo?",
			CreatedAt:      time.Now().Add(-5 * time.Minute),
		},
	}); err != nil {
		t.Fatalf("setup SaveMessages error: %v", err)
	}

	var messagesReceived []llm.Message
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			messagesReceived = in.Messages
			return llm.ChatCompletionOutput{Content: "Tidak ada promo saat ini."}, nil
		},
	}

	service := NewAssistantService(repo, model, productService, convRepo)

	_, err := service.ChatWithConversation(context.Background(), "Baik, terima kasih.", convID, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that the orphan tool message was stripped and never sent to LLM
	for _, m := range messagesReceived {
		if m.Role == "tool" && m.ToolCallID == "call_xyz" {
			t.Fatal("orphan tool message was not sanitized from LLM prompt")
		}
	}
}

func embeddingFixture() []float32 {
	return make([]float32, constants.AssistantEmbeddingDimension)
}

type fakeApplicationService struct {
	createdApp    models.Application
	capturedSlug  string
	capturedInput dtos.CreateApplicationRequest
	err           error
}

func (f *fakeApplicationService) Create(ctx context.Context, slug string, input dtos.CreateApplicationRequest) (models.Application, error) {
	if f.err != nil {
		return models.Application{}, f.err
	}
	f.capturedSlug = slug
	f.capturedInput = input
	return f.createdApp, nil
}

func TestAssistantChatWithTools_SubmitApplication_Success(t *testing.T) {
	repo := &knowledgeFake{}
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)
	appService := &fakeApplicationService{
		createdApp: models.Application{
			ID:        "app-xyz-123",
			ProductID: "prod-1",
			FullName:  "John Doe",
			Premium:   500000,
			Status:    models.ApplicationStatusSubmitted,
		},
	}

	callCount := 0
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			callCount++
			if callCount == 1 {
				return llm.ChatCompletionOutput{
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_submit_1",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name: "submit_application",
								Arguments: `{
									"product_slug": "secure-life-plus",
									"full_name": "John Doe",
									"email": "john.doe@example.com",
									"phone": "0812-3456-7890",
									"age": 30,
									"gender": "male",
									"sum_assured": 500000000,
									"payment_term": 10,
									"payment_frequency": "monthly"
								}`,
							},
						},
					},
				}, nil
			}
			return llm.ChatCompletionOutput{
				Content: "Pengajuan Anda telah berhasil diserahkan dengan Nomor Pengajuan app-xyz-123.",
			}, nil
		},
	}

	service := NewAssistantService(repo, model, productService).WithApplicationService(appService)

	resp, err := service.Chat(context.Background(), "Saya konfirmasi setuju mendaftar asuransi.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp.Answer, "app-xyz-123") {
		t.Fatalf("expected answer to contain app-xyz-123, got: %s", resp.Answer)
	}
	if len(resp.ToolsUsed) != 1 || resp.ToolsUsed[0] != "submit_application" {
		t.Fatalf("expected tools_used to be [submit_application], got: %v", resp.ToolsUsed)
	}
	if appService.capturedSlug != "secure-life-plus" {
		t.Fatalf("expected slug secure-life-plus, got: %s", appService.capturedSlug)
	}
	if appService.capturedInput.FullName != "John Doe" {
		t.Fatalf("expected FullName John Doe, got: %s", appService.capturedInput.FullName)
	}
	if appService.capturedInput.Email != "john.doe@example.com" {
		t.Fatalf("expected Email john.doe@example.com, got: %s", appService.capturedInput.Email)
	}
	if appService.capturedInput.Phone != "081234567890" {
		t.Fatalf("expected sanitized Phone 081234567890, got: %s", appService.capturedInput.Phone)
	}
}

func TestAssistantChatWithTools_SubmitApplication_ValidationError(t *testing.T) {
	repo := &knowledgeFake{}
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)
	appService := &fakeApplicationService{}

	callCount := 0
	var receivedToolResult string
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			callCount++
			if callCount == 1 {
				return llm.ChatCompletionOutput{
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_submit_invalid",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name: "submit_application",
								Arguments: `{
									"product_slug": "secure-life-plus",
									"full_name": "John Doe",
									"email": "invalid-email-address",
									"phone": "08123456789",
									"age": 30,
									"gender": "male",
									"sum_assured": 500000000,
									"payment_term": 10
								}`,
							},
						},
					},
				}, nil
			}
			for _, m := range in.Messages {
				if m.Role == "tool" {
					receivedToolResult = m.Content
				}
			}
			return llm.ChatCompletionOutput{
				Content: "Format email Anda tidak valid, mohon berikan email yang benar.",
			}, nil
		},
	}

	service := NewAssistantService(repo, model, productService).WithApplicationService(appService)

	resp, err := service.Chat(context.Background(), "Tolong daftarkan saya.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedToolResult, constants.ErrApplicationEmailInvalid) {
		t.Fatalf("expected tool result to report email invalid, got: %s", receivedToolResult)
	}
	if !strings.Contains(resp.Answer, "email") {
		t.Fatalf("expected assistant response to mention email, got: %s", resp.Answer)
	}
}

func TestAssistantChatWithTools_SubmitApplication_ServiceUnavailable(t *testing.T) {
	repo := &knowledgeFake{}
	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	callCount := 0
	var receivedToolResult string
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			callCount++
			if callCount == 1 {
				return llm.ChatCompletionOutput{
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_submit_unavailable",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name: "submit_application",
								Arguments: `{
									"product_slug": "secure-life-plus",
									"full_name": "John Doe",
									"email": "john.doe@example.com",
									"phone": "08123456789",
									"age": 30,
									"gender": "male",
									"sum_assured": 500000000,
									"payment_term": 10
								}`,
							},
						},
					},
				}, nil
			}
			for _, m := range in.Messages {
				if m.Role == "tool" {
					receivedToolResult = m.Content
				}
			}
			return llm.ChatCompletionOutput{
				Content: "Layanan pendaftaran sedang tidak tersedia saat ini.",
			}, nil
		},
	}

	// Service without application service wired
	service := NewAssistantService(repo, model, productService)

	resp, err := service.Chat(context.Background(), "Daftar sekarang.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedToolResult, "application service is unavailable") {
		t.Fatalf("expected tool result to report service unavailable, got: %s", receivedToolResult)
	}
	if !strings.Contains(resp.Answer, "tidak tersedia") {
		t.Fatalf("expected assistant response to indicate unavailable, got: %s", resp.Answer)
	}
}

func TestAssistantChatStream_Success(t *testing.T) {
	repo := &knowledgeFake{
		matches: []repositories.KnowledgeChunkMatch{
			{
				KnowledgeChunk: models.KnowledgeChunk{
					Title:      "FAQ Produk",
					Content:    "Produk asuransi jiwa mencakup santunan tutup usia.",
					SourceType: "faq",
				},
				Distance: 0.1,
			},
		},
	}
	model := &assistantLLMFake{
		embedding:    embeddingFixture(),
		streamTokens: []string{"Asuransi ", "jiwa ", "melindungi ", "keluarga ", "Anda."},
	}

	service := NewAssistantService(repo, model, nil)

	var receivedEvents []dtos.AssistantStreamEvent
	err := service.ChatStream(context.Background(), dtos.AssistantChatRequest{
		Message: "Jelaskan apa itu asuransi jiwa?",
	}, func(event dtos.AssistantStreamEvent) error {
		receivedEvents = append(receivedEvents, event)
		return nil
	})

	if err != nil {
		t.Fatalf("ChatStream failed: %v", err)
	}

	if len(receivedEvents) == 0 {
		t.Fatal("expected to receive stream events, got none")
	}

	var tokenContent strings.Builder
	var doneReceived bool
	for _, ev := range receivedEvents {
		switch ev.Type {
		case dtos.StreamEventToken:
			tokenContent.WriteString(ev.Content)
		case dtos.StreamEventDone:
			doneReceived = true
			if ev.ConversationID == "" {
				t.Errorf("expected conversation_id in done event")
			}
			if len(ev.Sources) != 1 {
				t.Errorf("expected 1 source in done event, got %d", len(ev.Sources))
			}
		}
	}

	if !doneReceived {
		t.Fatal("expected done event")
	}
	expectedText := "Asuransi jiwa melindungi keluarga Anda."
	if tokenContent.String() != expectedText {
		t.Fatalf("expected streamed content %q, got %q", expectedText, tokenContent.String())
	}
}

func TestAssistantChatStream_WithTools(t *testing.T) {
	repo := &knowledgeFake{
		matches: []repositories.KnowledgeChunkMatch{
			{
				KnowledgeChunk: models.KnowledgeChunk{
					Title:      "Produk",
					Content:    "List of products",
					SourceType: "product",
				},
				Distance: 0.1,
			},
		},
	}

	products := &fakeProductRepository{product: productFixture()}
	productService := NewProductService(products)

	toolCallsMade := false
	model := &assistantLLMFake{
		embedding: embeddingFixture(),
		chatWithToolsFn: func(ctx context.Context, in llm.ChatCompletionInput) (llm.ChatCompletionOutput, error) {
			if !toolCallsMade {
				toolCallsMade = true
				return llm.ChatCompletionOutput{
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_quote_1",
							Type: "function",
							Function: llm.ToolCallFunction{
								Name:      "calculate_quote",
								Arguments: `{"product_slug":"secure-life-plus","age":30,"gender":"male","sum_assured":100000000,"payment_term":10,"payment_frequency":"monthly"}`,
							},
						},
					},
				}, nil
			}
			return llm.ChatCompletionOutput{}, nil
		},
		streamTokens: []string{"Premi ", "tahunan ", "Anda ", "adalah ", "Rp1.000.000."},
	}

	service := NewAssistantService(repo, model, productService)

	var receivedEvents []dtos.AssistantStreamEvent
	err := service.ChatStream(context.Background(), dtos.AssistantChatRequest{
		Message: "Berapa premi saya?",
	}, func(event dtos.AssistantStreamEvent) error {
		receivedEvents = append(receivedEvents, event)
		return nil
	})

	if err != nil {
		t.Fatalf("ChatStream with tools failed: %v", err)
	}

	var hasToolCall, hasToolResult, hasDone bool
	var toolsUsedInDone []string
	var tokenContent strings.Builder

	for _, ev := range receivedEvents {
		switch ev.Type {
		case dtos.StreamEventToolCall:
			hasToolCall = true
			if ev.ToolName != "calculate_quote" {
				t.Errorf("expected tool_name calculate_quote, got %s", ev.ToolName)
			}
		case dtos.StreamEventToolResult:
			hasToolResult = true
			if ev.ToolName != "calculate_quote" {
				t.Errorf("expected tool_name calculate_quote, got %s", ev.ToolName)
			}
		case dtos.StreamEventToken:
			tokenContent.WriteString(ev.Content)
		case dtos.StreamEventDone:
			hasDone = true
			toolsUsedInDone = ev.ToolsUsed
		}
	}

	if !hasToolCall {
		t.Error("expected tool_call event")
	}
	if !hasToolResult {
		t.Error("expected tool_result event")
	}
	if !hasDone {
		t.Error("expected done event")
	}
	if len(toolsUsedInDone) != 1 || toolsUsedInDone[0] != "calculate_quote" {
		t.Errorf("expected tools_used [calculate_quote], got %v", toolsUsedInDone)
	}
	if tokenContent.String() != "Premi tahunan Anda adalah Rp1.000.000." {
		t.Errorf("unexpected token content: %s", tokenContent.String())
	}
}

func TestAssistantChatStream_ValidationErrors(t *testing.T) {
	service := NewAssistantService(nil, nil, nil)
	err := service.ChatStream(context.Background(), dtos.AssistantChatRequest{
		Message: "",
	}, nil)
	if !errors.Is(err, constants.ErrAssistantMessageRequiredError) {
		t.Fatalf("expected ErrAssistantMessageRequiredError, got %v", err)
	}
}
