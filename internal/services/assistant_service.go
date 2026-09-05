package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	pgvector "github.com/pgvector/pgvector-go"
)

type AssistantService struct {
	knowledgeRepository repositories.KnowledgeRepository
	llm                 AssistantLLM
	quotes              AssistantQuoteService
}

type AssistantLLM interface {
	CreateChatCompletion(context.Context, llm.ChatCompletionInput) (string, error)
	CreateEmbedding(context.Context, llm.EmbeddingInput) ([]float32, error)
}

type AssistantQuoteService interface {
	CreateProductQuote(context.Context, string, dtos.CreateProductQuoteInput) (dtos.ProductQuote, error)
}

func (service *AssistantService) Chat(ctx context.Context, message string) (dtos.AssistantChatResponse, error) {
	return service.chat(ctx, message, nil, "")
}

func (service *AssistantService) ChatWithQuote(ctx context.Context, message string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error) {
	return service.chat(ctx, message, quote, slug)
}

func NewAssistantService(knowledgeRepository repositories.KnowledgeRepository, llmClient AssistantLLM, quoteService AssistantQuoteService) *AssistantService {
	return &AssistantService{knowledgeRepository: knowledgeRepository, llm: llmClient, quotes: quoteService}
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

func (service *AssistantService) chat(ctx context.Context, message string, quoteRequest *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error) {
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
	answer, err := service.llm.CreateChatCompletion(ctx, llm.ChatCompletionInput{
		Messages: []llm.Message{
			{Role: "system", Content: "You are an insurance assistant. Answer using only the provided context. If the context is insufficient, say you do not know."},
			{Role: "user", Content: "Context:\n" + contextText + "\n\nQuestion: " + message},
		},
	})
	if err != nil {
		return dtos.AssistantChatResponse{}, err
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

	return dtos.AssistantChatResponse{Answer: answer, Sources: sources}, nil
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
