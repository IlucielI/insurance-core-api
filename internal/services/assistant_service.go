package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	pgvector "github.com/pgvector/pgvector-go"
)

type AssistantService struct {
	knowledgeRepository    repositories.KnowledgeRepository
	llm                    AssistantLLM
	quotes                 AssistantQuoteService
	conversationRepository repositories.AssistantConversationRepository
}

type AssistantLLM interface {
	CreateChatCompletion(context.Context, llm.ChatCompletionInput) (string, error)
	CreateChatCompletionWithTools(context.Context, llm.ChatCompletionInput) (llm.ChatCompletionOutput, error)
	CreateEmbedding(context.Context, llm.EmbeddingInput) ([]float32, error)
}

type AssistantQuoteService interface {
	CreateProductQuote(context.Context, string, dtos.CreateProductQuoteInput) (dtos.ProductQuote, error)
}

type AssistantProductLister interface {
	ListProducts(context.Context, dtos.ProductListQuery) ([]models.Product, error)
}

func (service *AssistantService) Chat(ctx context.Context, message string) (dtos.AssistantChatResponse, error) {
	return service.chat(ctx, message, nil, "", "")
}

func (service *AssistantService) ChatWithQuote(ctx context.Context, message string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error) {
	return service.chat(ctx, message, quote, slug, "")
}

func (service *AssistantService) ChatWithConversation(ctx context.Context, message string, conversationID string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error) {
	return service.chat(ctx, message, quote, slug, conversationID)
}

func NewAssistantService(knowledgeRepository repositories.KnowledgeRepository, llmClient AssistantLLM, quoteService AssistantQuoteService, conversationRepository ...repositories.AssistantConversationRepository) *AssistantService {
	var convRepo repositories.AssistantConversationRepository
	if len(conversationRepository) > 0 {
		convRepo = conversationRepository[0]
	}
	return &AssistantService{
		knowledgeRepository:    knowledgeRepository,
		llm:                    llmClient,
		quotes:                 quoteService,
		conversationRepository: convRepo,
	}
}

func (service *AssistantService) GetConversation(ctx context.Context, conversationID string) (dtos.AssistantConversationResponse, error) {
	if service.conversationRepository == nil {
		return dtos.AssistantConversationResponse{}, constants.ErrAssistantServiceUnavailableError
	}
	conv, err := service.conversationRepository.GetConversation(ctx, conversationID)
	if err != nil {
		return dtos.AssistantConversationResponse{}, err
	}
	msgResponses := make([]dtos.AssistantMessageResponse, 0, len(conv.Messages))
	for _, m := range conv.Messages {
		msgResponses = append(msgResponses, dtos.AssistantMessageResponse{
			ID:         m.ID,
			Role:       m.Role,
			Content:    m.Content,
			ToolCalls:  m.ToolCalls,
			ToolCallID: m.ToolCallID,
			CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		})
	}
	return dtos.AssistantConversationResponse{
		ID:        conv.ID,
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt.Format(time.RFC3339),
		UpdatedAt: conv.UpdatedAt.Format(time.RFC3339),
		Messages:  msgResponses,
	}, nil
}

func (service *AssistantService) DeleteConversation(ctx context.Context, conversationID string) error {
	if service.conversationRepository == nil {
		return constants.ErrAssistantServiceUnavailableError
	}
	return service.conversationRepository.DeleteConversation(ctx, conversationID)
}

func (service *AssistantService) SeedDefaultKnowledge(ctx context.Context) error {
	if service.knowledgeRepository == nil || service.llm == nil {
		return constants.ErrAssistantServiceUnavailableError
	}
	count, err := service.knowledgeRepository.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	chunks := chunkKnowledge(defaultKnowledgeChunks())
	for index := range chunks {
		embedding, err := service.llm.CreateEmbedding(ctx, llm.EmbeddingInput{Text: chunks[index].Content})
		if err != nil {
			return fmt.Errorf("embed knowledge chunk %q: %w", chunks[index].ID, err)
		}
		if len(embedding) != constants.AssistantEmbeddingDimension {
			return fmt.Errorf("embed knowledge chunk %q: embedding dimension must be %d", chunks[index].ID, constants.AssistantEmbeddingDimension)
		}
		chunks[index].Embedding = pgvector.NewVector(embedding)
	}

	return service.knowledgeRepository.ReplaceAll(ctx, chunks)
}

func (service *AssistantService) chat(ctx context.Context, message string, quoteRequest *dtos.ProductQuoteRequest, slug string, conversationID string) (dtos.AssistantChatResponse, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return dtos.AssistantChatResponse{}, constants.ErrAssistantMessageRequiredError
	}
	if len(message) > constants.AssistantMaxMessageSize {
		return dtos.AssistantChatResponse{}, constants.ErrAssistantMessageTooLongError
	}
	if service.knowledgeRepository == nil || service.llm == nil {
		return dtos.AssistantChatResponse{}, constants.ErrAssistantServiceUnavailableError
	}

	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		conversationID = generateID()
	}

	var historyMessages []llm.Message
	if service.conversationRepository != nil {
		_, _ = service.conversationRepository.GetOrCreateConversation(ctx, conversationID, "Insurance Consultation")
		history, err := service.conversationRepository.ListMessages(ctx, conversationID, 10)
		if err == nil {
			for _, h := range history {
				historyMessages = append(historyMessages, llm.Message{
					Role:       h.Role,
					Content:    h.Content,
					ToolCalls:  h.ToolCalls,
					ToolCallID: h.ToolCallID,
				})
			}
		}
	}

	embedding, err := service.llm.CreateEmbedding(ctx, llm.EmbeddingInput{Text: message})
	if err != nil {
		return dtos.AssistantChatResponse{}, err
	}

	matches, err := service.knowledgeRepository.Search(ctx, embedding, constants.AssistantTopK)
	if err != nil {
		return dtos.AssistantChatResponse{}, err
	}

	filtered := matches[:0]
	for _, match := range matches {
		if match.Distance <= constants.AssistantMaxDistance {
			filtered = append(filtered, match)
		}
	}
	matches = filtered
	quoteContext, err := service.buildQuoteContext(ctx, quoteRequest, slug)
	if err != nil {
		return dtos.AssistantChatResponse{}, err
	}

	contextText := strings.TrimSpace(buildContext(matches) + "\n\n" + quoteContext)
	systemMsg := llm.Message{
		Role:    "system",
		Content: "You are an insurance assistant. Answer using only the provided context, conversation history, and tools. If the context is insufficient, say you do not know. If the user wants to calculate premium and provides the required details, call the calculate_quote tool. If required information is missing, ask the user to provide it. If the user asks about available products, call list_products.",
	}
	userMsg := llm.Message{
		Role:    "user",
		Content: strings.TrimSpace("Context:\n" + contextText + "\n\nQuestion: " + message),
	}

	messages := make([]llm.Message, 0, len(historyMessages)+2)
	messages = append(messages, systemMsg)
	messages = append(messages, historyMessages...)
	messages = append(messages, userMsg)

	newMessages := []models.AssistantMessage{
		{
			ID:             generateID(),
			ConversationID: conversationID,
			Role:           "user",
			Content:        message,
			CreatedAt:      time.Now(),
		},
	}

	tools := assistantTools()
	toolsUsed := make([]string, 0)
	var answer string

	for iteration := 0; iteration < 3; iteration++ {
		output, err := service.llm.CreateChatCompletionWithTools(ctx, llm.ChatCompletionInput{
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return dtos.AssistantChatResponse{}, err
		}

		if len(output.ToolCalls) == 0 {
			answer = output.Content
			break
		}

		messages = append(messages, llm.Message{
			Role:      "assistant",
			Content:   output.Content,
			ToolCalls: output.ToolCalls,
		})

		newMessages = append(newMessages, models.AssistantMessage{
			ID:             generateID(),
			ConversationID: conversationID,
			Role:           "assistant",
			Content:        output.Content,
			ToolCalls:      output.ToolCalls,
			CreatedAt:      time.Now(),
		})

		for _, toolCall := range output.ToolCalls {
			toolsUsed = append(toolsUsed, toolCall.Function.Name)
			toolResult := service.executeTool(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: toolCall.ID,
			})
			newMessages = append(newMessages, models.AssistantMessage{
				ID:             generateID(),
				ConversationID: conversationID,
				Role:           "tool",
				Content:        toolResult,
				ToolCallID:     toolCall.ID,
				CreatedAt:      time.Now(),
			})
		}
	}

	if answer == "" {
		fallback, err := service.llm.CreateChatCompletion(ctx, llm.ChatCompletionInput{
			Messages: messages,
		})
		if err == nil {
			answer = fallback
		}
	}

	if answer != "" {
		newMessages = append(newMessages, models.AssistantMessage{
			ID:             generateID(),
			ConversationID: conversationID,
			Role:           "assistant",
			Content:        answer,
			CreatedAt:      time.Now(),
		})
	}

	if service.conversationRepository != nil && len(newMessages) > 0 {
		_ = service.conversationRepository.SaveMessages(ctx, newMessages)
	}

	sources := make([]dtos.AssistantSource, 0, len(matches))
	for _, match := range matches {
		sources = append(sources, dtos.AssistantSource{
			Title:      match.Title,
			SourceType: match.SourceType,
			Score:      max(0, 1-match.Distance),
			Excerpt:    excerpt(match.Content),
		})
	}

	return dtos.AssistantChatResponse{
		ConversationID: conversationID,
		Answer:         answer,
		Sources:        sources,
		ToolsUsed:      toolsUsed,
	}, nil
}

func generateID() string {
	value := make([]byte, 16)
	_, _ = rand.Read(value)
	return hex.EncodeToString(value)
}

func (service *AssistantService) executeTool(ctx context.Context, name string, rawArgs string) string {
	switch name {
	case "calculate_quote":
		if service.quotes == nil {
			return `{"error":"quote service is unavailable"}`
		}
		var args struct {
			ProductSlug      string `json:"product_slug"`
			Age              int    `json:"age"`
			Gender           string `json:"gender"`
			SumAssured       int64  `json:"sum_assured"`
			PaymentTerm      int    `json:"payment_term"`
			PaymentFrequency string `json:"payment_frequency"`
			Smoker           string `json:"smoker"`
			OccupationClass  string `json:"occupation_class"`
			HealthRisk       string `json:"health_risk"`
		}
		if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
			return fmt.Sprintf(`{"error":"invalid arguments: %s"}`, err.Error())
		}
		args.ProductSlug = strings.TrimSpace(args.ProductSlug)
		if args.ProductSlug == "" {
			return `{"error":"product_slug is required"}`
		}

		quoteReq := dtos.ProductQuoteRequest{
			Age:              args.Age,
			Gender:           strings.ToLower(strings.TrimSpace(args.Gender)),
			SumAssured:       args.SumAssured,
			PaymentTerm:      args.PaymentTerm,
			PaymentFrequency: strings.ToLower(strings.TrimSpace(args.PaymentFrequency)),
			Smoker:           strings.ToLower(strings.TrimSpace(args.Smoker)),
			OccupationClass:  strings.ToLower(strings.TrimSpace(args.OccupationClass)),
			HealthRisk:       strings.ToLower(strings.TrimSpace(args.HealthRisk)),
		}
		if quoteReq.PaymentFrequency == "" {
			quoteReq.PaymentFrequency = constants.PaymentFrequencyMonthly
		}
		if quoteReq.Smoker == "" || quoteReq.Smoker == "non_smoker" {
			quoteReq.Smoker = constants.SmokerNo
		} else if quoteReq.Smoker == "smoker" {
			quoteReq.Smoker = constants.SmokerYes
		}
		if quoteReq.OccupationClass == "" || quoteReq.OccupationClass == "1" || quoteReq.OccupationClass == "standard" {
			quoteReq.OccupationClass = constants.OccupationStandard
		} else if quoteReq.OccupationClass == "low" {
			quoteReq.OccupationClass = constants.OccupationLow
		} else if quoteReq.OccupationClass == "high" {
			quoteReq.OccupationClass = constants.OccupationHigh
		}
		if quoteReq.HealthRisk == "" || quoteReq.HealthRisk == "standard" || quoteReq.HealthRisk == "low" {
			quoteReq.HealthRisk = constants.HealthRiskLow
		}

		validatedReq, err := validations.ValidateProductQuoteRequest(quoteReq)
		if err != nil {
			return fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}

		quote, err := service.quotes.CreateProductQuote(ctx, args.ProductSlug, dtos.ProductQuoteRequestToInput(validatedReq))
		if err != nil {
			return fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}
		res, err := json.Marshal(quote)
		if err != nil {
			return fmt.Sprintf(`{"error":"failed to serialize quote: %s"}`, err.Error())
		}
		return string(res)

	case "list_products":
		lister, ok := service.quotes.(AssistantProductLister)
		if !ok || lister == nil {
			return `{"error":"product list service is unavailable"}`
		}
		var args struct {
			Category string `json:"category"`
			Search   string `json:"search"`
		}
		if len(strings.TrimSpace(rawArgs)) > 0 {
			if err := json.Unmarshal([]byte(rawArgs), &args); err != nil {
				return fmt.Sprintf(`{"error":"invalid arguments: %s"}`, err.Error())
			}
		}
		products, err := lister.ListProducts(ctx, dtos.ProductListQuery{
			Category: strings.TrimSpace(args.Category),
			Search:   strings.TrimSpace(args.Search),
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"%s"}`, err.Error())
		}
		type itemSummary struct {
			Name          string `json:"name"`
			Slug          string `json:"slug"`
			Category      string `json:"category"`
			MinSumAssured int64  `json:"min_sum_assured"`
			MaxSumAssured int64  `json:"max_sum_assured"`
			MinTerm       int    `json:"min_payment_term"`
			MaxTerm       int    `json:"max_payment_term"`
		}
		items := make([]itemSummary, 0, len(products))
		for _, p := range products {
			items = append(items, itemSummary{
				Name:          p.Name,
				Slug:          p.Slug,
				Category:      string(p.Category),
				MinSumAssured: p.MinSumAssured,
				MaxSumAssured: p.MaxSumAssured,
				MinTerm:       p.MinPaymentTerm,
				MaxTerm:       p.MaxPaymentTerm,
			})
		}
		res, err := json.Marshal(items)
		if err != nil {
			return fmt.Sprintf(`{"error":"failed to serialize products: %s"}`, err.Error())
		}
		return string(res)

	default:
		return fmt.Sprintf(`{"error":"unknown tool %s"}`, name)
	}
}

func assistantTools() []llm.Tool {
	return []llm.Tool{
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "calculate_quote",
				Description: "Calculates estimated insurance premium for a product based on user parameters.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"product_slug": map[string]any{
							"type":        "string",
							"description": "The slug of the product to quote, e.g. 'secure-life-plus'",
						},
						"age": map[string]any{
							"type":        "integer",
							"description": "Age of the insured in years",
						},
						"gender": map[string]any{
							"type":        "string",
							"enum":        []string{"male", "female"},
							"description": "Gender of the insured ('male' or 'female')",
						},
						"sum_assured": map[string]any{
							"type":        "integer",
							"description": "Sum assured / coverage amount in IDR",
						},
						"payment_term": map[string]any{
							"type":        "integer",
							"description": "Payment term duration in years",
						},
						"payment_frequency": map[string]any{
							"type":        "string",
							"enum":        []string{"monthly", "annual", "quarterly", "semi_annual"},
							"description": "Payment frequency (default is 'monthly')",
						},
						"smoker": map[string]any{
							"type":        "string",
							"enum":        []string{"smoker", "non_smoker"},
							"description": "Smoker status",
						},
						"occupation_class": map[string]any{
							"type":        "string",
							"description": "Occupation class ('1', '2', '3', '4')",
						},
						"health_risk": map[string]any{
							"type":        "string",
							"enum":        []string{"standard", "substandard"},
							"description": "Health risk status",
						},
					},
					"required": []string{"product_slug", "age", "gender", "sum_assured", "payment_term"},
				},
			},
		},
		{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        "list_products",
				Description: "Lists or searches available insurance products and their coverage details.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"category": map[string]any{
							"type":        "string",
							"description": "Product category to filter by (e.g. 'life', 'health')",
						},
						"search": map[string]any{
							"type":        "string",
							"description": "Search keyword for product name",
						},
					},
				},
			},
		},
	}
}

func (service *AssistantService) buildQuoteContext(ctx context.Context, quoteRequest *dtos.ProductQuoteRequest, slug string) (string, error) {
	if quoteRequest == nil {
		return "", nil
	}
	if service.quotes == nil {
		return "", constants.ErrAssistantServiceUnavailableError
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "", errors.New("product slug is required for quote calculation")
	}
	quote, err := service.quotes.CreateProductQuote(ctx, slug, dtos.ProductQuoteRequestToInput(*quoteRequest))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[quote_calculation] Product: %s. Sum assured: %d %s. Payment term: %d years. Payment frequency: %s. Estimated premium: %d %s. Estimated annual premium: %d %s.", quote.ProductName, quote.SumAssured, quote.Currency, quote.PaymentTerm, quote.PaymentFrequency, quote.EstimatedPremium, quote.Currency, quote.EstimatedAnnualPremium, quote.Currency), nil
}

func defaultKnowledgeChunks() []models.KnowledgeChunk {
	return []models.KnowledgeChunk{
		{ID: "knowledge-product-summary", SourceType: constants.AssistantSourceTypeProduct, Title: "Product Overview", Content: "The app offers insurance products with product details, benefits, exclusions, premium starting points, and quote calculator support. Users can browse products, see details, and generate indicative quotes before applying."},
		{ID: "knowledge-quote-flow", SourceType: constants.AssistantSourceTypeQuote, Title: "Quote Calculator", Content: "Premium quotes are calculated from age, gender, sum assured, payment term, payment frequency, smoker status, occupation class, and health risk. The quote is indicative and final premium can change after underwriting review."},
		{ID: "knowledge-application-flow", SourceType: constants.AssistantSourceTypeApplicationFlow, Title: "Application Flow", Content: "A customer starts by selecting a product, then creates a quote, submits an application, and the application moves through submitted, under_review, approved, or rejected statuses."},
		{ID: "knowledge-underwriting-flow", SourceType: constants.AssistantSourceTypeUnderwriting, Title: "Underwriting Checklist", Content: "Underwriters review identity verification, income verification, documents completeness, and medical requirements. An application can only be approved after required review checks are passed or marked not needed."},
		{ID: "knowledge-company-info", SourceType: constants.AssistantSourceTypeCompany, Title: "Company Information", Content: "The insurance core API is a backend service for insurance policy applications. Company profile, address, and official public details can be added later from a CMS or admin source."},
		{ID: "knowledge-faq", SourceType: constants.AssistantSourceTypeFAQ, Title: "FAQ", Content: "If the answer depends on a live calculation or current application data, the assistant should rely on backend services. If the answer is not present in the available knowledge, it should say it does not know."},
	}
}

func buildContext(matches []repositories.KnowledgeChunkMatch) string {
	parts := make([]string, 0, len(matches))
	for _, match := range matches {
		parts = append(parts, fmt.Sprintf("[%s] %s: %s", match.SourceType, match.Title, match.Content))
	}
	return strings.Join(parts, "\n\n")
}

func excerpt(content string) string {
	if len(content) <= 180 {
		return content
	}
	return content[:180] + "..."
}

func chunkKnowledge(chunks []models.KnowledgeChunk) []models.KnowledgeChunk {
	result := make([]models.KnowledgeChunk, 0, len(chunks))
	for _, source := range chunks {
		words := strings.Fields(source.Content)
		if len(words) == 0 {
			continue
		}
		for start, index := 0, 0; start < len(words); index++ {
			end := start + constants.AssistantChunkTokenLimit
			if end > len(words) {
				end = len(words)
			}
			chunk := source
			chunk.ID = fmt.Sprintf("%s-%d", source.ID, index)
			chunk.ChunkIndex = index
			chunk.Content = strings.Join(words[start:end], " ")
			result = append(result, chunk)
			if end == len(words) {
				break
			}
			start = end - constants.AssistantChunkOverlap
		}
	}
	return result
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
