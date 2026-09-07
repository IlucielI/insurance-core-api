package services

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/models"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	pgvector "github.com/pgvector/pgvector-go"
)

type KnowledgeRAGLLM interface {
	CreateEmbedding(context.Context, llm.EmbeddingInput) ([]float32, error)
	CreateChatCompletion(context.Context, llm.ChatCompletionInput) (string, error)
}

type KnowledgeRAGService interface {
	ReindexDocument(ctx context.Context, id string) (dtos.ReindexDocumentResponse, error)
	GetMetrics(ctx context.Context) (dtos.KnowledgeBaseMetricsResponse, error)
	SimulateChat(ctx context.Context, req dtos.SimulateChatRequest) (dtos.SimulateChatResponse, error)
	SyncDocumentChunks(ctx context.Context, doc models.KnowledgeDocument) (int, error)
}

type DefaultKnowledgeRAGService struct {
	docRepo     repositories.KnowledgeDocumentRepository
	chunkRepo   repositories.KnowledgeRepository
	metricsRepo repositories.KnowledgeMetricsRepository
	llm         KnowledgeRAGLLM
}

func NewKnowledgeRAGService(
	docRepo repositories.KnowledgeDocumentRepository,
	chunkRepo repositories.KnowledgeRepository,
	metricsRepo repositories.KnowledgeMetricsRepository,
	llmClient KnowledgeRAGLLM,
) *DefaultKnowledgeRAGService {
	return &DefaultKnowledgeRAGService{
		docRepo:     docRepo,
		chunkRepo:   chunkRepo,
		metricsRepo: metricsRepo,
		llm:         llmClient,
	}
}

func (s *DefaultKnowledgeRAGService) embedChunk(ctx context.Context, text string) ([]float32, error) {
	if s.llm != nil {
		emb, err := s.llm.CreateEmbedding(ctx, llm.EmbeddingInput{Text: text})
		if err == nil && len(emb) == constants.AssistantEmbeddingDimension {
			return emb, nil
		}
	}
	return generateDeterministicEmbedding(text, constants.AssistantEmbeddingDimension), nil
}

func (s *DefaultKnowledgeRAGService) SyncDocumentChunks(ctx context.Context, doc models.KnowledgeDocument) (int, error) {
	chunks := ChunkKnowledgeDocument(doc)
	if len(chunks) == 0 {
		if s.chunkRepo != nil {
			_ = s.chunkRepo.ReplaceByDocumentID(ctx, doc.ID, nil)
		}
		return 0, nil
	}

	for i := range chunks {
		emb, err := s.embedChunk(ctx, chunks[i].Content)
		if err != nil {
			return 0, err
		}
		chunks[i].Embedding = pgvector.NewVector(emb)
	}

	if s.chunkRepo != nil {
		if err := s.chunkRepo.ReplaceByDocumentID(ctx, doc.ID, chunks); err != nil {
			return 0, err
		}
	}

	now := time.Now().UTC()
	doc.ChunkCount = len(chunks)
	doc.Status = models.IndexingStatusIndexed
	doc.LastSyncedAt = &now
	doc.UpdatedAt = now

	if s.docRepo != nil {
		_ = s.docRepo.Update(ctx, doc.ID, &doc)
	}

	return len(chunks), nil
}

func (s *DefaultKnowledgeRAGService) ReindexDocument(ctx context.Context, id string) (dtos.ReindexDocumentResponse, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return dtos.ReindexDocumentResponse{}, errors.New(constants.ErrKnowledgeDocIDRequired)
	}

	if s.docRepo == nil {
		return dtos.ReindexDocumentResponse{}, errors.New("document repository not configured")
	}

	doc, err := s.docRepo.FindByID(ctx, trimmedID)
	if err != nil {
		if docBySlug, errSlug := s.docRepo.FindBySlug(ctx, trimmedID); errSlug == nil {
			doc = docBySlug
		} else {
			return dtos.ReindexDocumentResponse{}, err
		}
	}

	// Update status to syncing
	doc.Status = models.IndexingStatusSyncing
	now := time.Now().UTC()
	doc.UpdatedAt = now
	_ = s.docRepo.Update(ctx, doc.ID, &doc)

	count, err := s.SyncDocumentChunks(ctx, doc)
	if err != nil {
		return dtos.ReindexDocumentResponse{}, err
	}

	return dtos.ReindexDocumentResponse{
		DocumentID:   doc.ID,
		Title:        doc.Title,
		Status:       string(models.IndexingStatusIndexed),
		ChunkCount:   count,
		LastSyncedAt: time.Now().UTC(),
	}, nil
}

func (s *DefaultKnowledgeRAGService) GetMetrics(ctx context.Context) (dtos.KnowledgeBaseMetricsResponse, error) {
	if s.metricsRepo == nil {
		return dtos.KnowledgeBaseMetricsResponse{}, errors.New("metrics repository not configured")
	}
	return s.metricsRepo.GetKnowledgeMetrics(ctx)
}

func (s *DefaultKnowledgeRAGService) SimulateChat(ctx context.Context, req dtos.SimulateChatRequest) (dtos.SimulateChatResponse, error) {
	start := time.Now()
	trimmedQuery := strings.TrimSpace(req.Query)
	if trimmedQuery == "" {
		return dtos.SimulateChatResponse{}, errors.New("query is required")
	}

	topK := req.TopK
	if topK <= 0 {
		topK = 3
	}
	if topK > 10 {
		topK = 10
	}

	emb, err := s.embedChunk(ctx, trimmedQuery)
	if err != nil {
		return dtos.SimulateChatResponse{}, err
	}

	if s.chunkRepo == nil {
		return dtos.SimulateChatResponse{}, errors.New("chunk repository not configured")
	}

	matches, err := s.chunkRepo.SearchWithCategory(ctx, emb, req.Category, topK)
	if err != nil {
		return dtos.SimulateChatResponse{}, err
	}

	retrievedChunks := make([]dtos.RetrievedChunkDetail, len(matches))
	contextParts := make([]string, 0, len(matches))

	for i, m := range matches {
		similarity := 1.0 - m.Distance
		if similarity < 0 {
			similarity = 0
		}
		if similarity > 1 {
			similarity = 1
		}
		retrievedChunks[i] = dtos.RetrievedChunkDetail{
			ID:         m.ID,
			DocumentID: m.DocumentID,
			Title:      m.Title,
			SourceType: m.SourceType,
			Content:    m.Content,
			Similarity: math.Round(similarity*1000) / 1000,
			ChunkIndex: m.ChunkIndex,
		}
		contextParts = append(contextParts, fmt.Sprintf("[%s] %s: %s", m.SourceType, m.Title, m.Content))
	}

	var answer string
	if len(retrievedChunks) == 0 {
		answer = "Maaf, informasi mengenai hal tersebut tidak ditemukan dalam basis data dokumen pengetahuan."
	} else if s.llm != nil {
		contextStr := strings.Join(contextParts, "\n\n")
		systemPrompt := "Anda adalah AI Copilot asuransi untuk CMS. Jawab pertanyaan pengguna secara akurat, ringkas, dan profesional hanya berdasarkan konteks dokumen SOP/regulasi yang diberikan. Jika konteks tidak memadai, jelaskan batasannya."
		prompt := fmt.Sprintf("Konteks Dokumen:\n%s\n\nPertanyaan: %s\n\nJawaban:", contextStr, trimmedQuery)

		temp := 0.2
		resp, err := s.llm.CreateChatCompletion(ctx, llm.ChatCompletionInput{
			Messages: []llm.Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: prompt},
			},
			Temperature: &temp,
		})
		if err == nil && resp != "" {
			answer = resp
		}
	}

	if answer == "" && len(retrievedChunks) > 0 {
		topMatch := retrievedChunks[0]
		answer = fmt.Sprintf("Berdasarkan dokumen '%s' (%s): %s", topMatch.Title, topMatch.SourceType, excerptContent(topMatch.Content))
	}

	return dtos.SimulateChatResponse{
		Query:           trimmedQuery,
		Answer:          answer,
		RetrievedChunks: retrievedChunks,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

func ChunkKnowledgeDocument(doc models.KnowledgeDocument) []models.KnowledgeChunk {
	words := strings.Fields(doc.Content)
	if len(words) == 0 {
		return nil
	}

	chunkLimit := constants.AssistantChunkTokenLimit
	if chunkLimit <= 0 {
		chunkLimit = 100
	}
	chunkOverlap := constants.AssistantChunkOverlap
	if chunkOverlap <= 0 {
		chunkOverlap = 20
	}

	result := make([]models.KnowledgeChunk, 0)
	for start, index := 0, 0; start < len(words); index++ {
		end := start + chunkLimit
		if end > len(words) {
			end = len(words)
		}

		chunkText := strings.Join(words[start:end], " ")
		chunkID := fmt.Sprintf("%s-chk-%d", doc.ID, index)

		chunk := models.KnowledgeChunk{
			ID:         chunkID,
			DocumentID: doc.ID,
			SourceType: string(doc.Category),
			Title:      fmt.Sprintf("%s (Bagian %d)", doc.Title, index+1),
			Content:    chunkText,
			ChunkIndex: index,
		}
		result = append(result, chunk)

		if end == len(words) {
			break
		}
		start = end - chunkOverlap
	}
	return result
}

func generateDeterministicEmbedding(text string, dim int) []float32 {
	vec := make([]float32, dim)
	h := sha256.Sum256([]byte(text))
	seed := binary.BigEndian.Uint64(h[:8])

	var sumSq float64
	for i := 0; i < dim; i++ {
		// Linear congruential pseudo-random float between -1.0 and 1.0
		seed = seed*6364136223846793005 + 1442695040888963407
		val := float32((float64(seed>>32)/float64(1<<32))*2.0 - 1.0)
		vec[i] = val
		sumSq += float64(val * val)
	}

	// L2 normalization
	norm := math.Sqrt(sumSq)
	if norm > 0 {
		for i := 0; i < dim; i++ {
			vec[i] = float32(float64(vec[i]) / norm)
		}
	}

	return vec
}

func excerptContent(s string) string {
	if len(s) <= 200 {
		return s
	}
	return s[:200] + "..."
}
