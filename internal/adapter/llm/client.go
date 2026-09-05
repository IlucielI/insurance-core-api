package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	BaseURL         string
	APIKey          string
	CompletionModel string
	EmbeddingModel  string
	Timeout         time.Duration
}

type Client struct {
	baseURL         string
	apiKey          string
	completionModel string
	embeddingModel  string
	httpClient      *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionInput struct {
	Messages    []Message
	Temperature *float64
	MaxTokens   *int
}

type EmbeddingInput struct {
	Text string
}

func NewClient(config Config) (*Client, error) {
	baseURL, err := normalizeBaseURL(config.BaseURL)
	if err != nil {
		return nil, err
	}

	completionModel := strings.TrimSpace(config.CompletionModel)
	if completionModel == "" {
		return nil, errors.New("CompletionModel is required")
	}

	embeddingModel := strings.TrimSpace(config.EmbeddingModel)
	if embeddingModel == "" {
		return nil, errors.New("EmbeddingModel is required")
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL:         baseURL,
		apiKey:          strings.TrimSpace(config.APIKey),
		completionModel: completionModel,
		embeddingModel:  embeddingModel,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (client *Client) CreateChatCompletion(ctx context.Context, input ChatCompletionInput) (string, error) {
	if len(input.Messages) == 0 {
		return "", errors.New("messages are required")
	}

	payload := chatCompletionRequest{
		Model:       client.completionModel,
		Messages:    input.Messages,
		Temperature: input.Temperature,
		MaxTokens:   input.MaxTokens,
	}

	raw, err := client.postRaw(ctx, "/chat/completions", payload)
	if err != nil {
		return "", err
	}

	return parseChatCompletionResponse(raw)
}

func (client *Client) CreateEmbedding(ctx context.Context, input EmbeddingInput) ([]float32, error) {
	if strings.TrimSpace(input.Text) == "" {
		return nil, errors.New("text is required")
	}

	payload := embeddingRequest{
		Model: client.embeddingModel,
		Input: input.Text,
	}

	raw, err := client.postRaw(ctx, "/embeddings", payload)
	if err != nil {
		return nil, err
	}

	return parseEmbeddingResponse(raw)
}

func (client *Client) post(ctx context.Context, path string, payload any, result any) error {
	raw, err := client.postRaw(ctx, path, payload)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, result)
}

func (client *Client) postRaw(ctx context.Context, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if client.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+client.apiKey)
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(response.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("llm request failed with status %d: %s", response.StatusCode, strings.TrimSpace(string(raw)))
	}

	return raw, nil
}

func normalizeBaseURL(rawBaseURL string) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(rawBaseURL), "/")
	if baseURL == "" {
		return "", errors.New("BaseURL is required")
	}

	parsedURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", errors.New("BaseURL must be a valid URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", errors.New("BaseURL scheme must be http or https")
	}

	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}

	return baseURL, nil
}

func parseChatCompletionResponse(raw []byte) (string, error) {
	var result chatCompletionResponse
	if err := json.Unmarshal(raw, &result); err == nil && len(result.Choices) > 0 {
		return parseContent(result.Choices[0].Message.Content)
	} else if err != nil && !looksLikeSSE(raw) {
		return "", fmt.Errorf("failed to parse chat completion response: %w", err)
	}

	return parseStreamingChatCompletionResponse(raw)
}

func looksLikeSSE(raw []byte) bool {
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "data:") {
			return true
		}
	}

	return false
}

func parseStreamingChatCompletionResponse(raw []byte) (string, error) {
	lines := strings.Split(string(raw), "\n")
	contents := make([]string, 0)
	var parseErrors []error

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}

		var chunk chatCompletionStreamResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			parseErrors = append(parseErrors, err)
			continue
		}

		for _, choice := range chunk.Choices {
			content, err := parseContent(choice.Delta.Content)
			if err == nil && content != "" {
				contents = append(contents, content)
			}
		}
	}

	if len(contents) == 0 && len(parseErrors) > 0 {
		return "", errors.Join(parseErrors...)
	}

	if len(contents) == 0 {
		return "", errors.New("streaming chat completion response has no content")
	}

	return strings.Join(contents, ""), nil
}

func parseContent(raw json.RawMessage) (string, error) {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if text == "" {
			return "", errors.New("chat completion content is empty")
		}

		return text, nil
	}

	var parts []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &parts); err == nil {
		values := make([]string, 0, len(parts))
		for _, part := range parts {
			if part.Text != "" {
				values = append(values, part.Text)
			}
		}

		if len(values) > 0 {
			return strings.Join(values, "\n"), nil
		}
	}

	return "", errors.New("chat completion content is empty or unsupported")
}

func parseEmbeddingResponse(raw []byte) ([]float32, error) {
	var openAIResponse struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &openAIResponse); err == nil && len(openAIResponse.Data) > 0 && len(openAIResponse.Data[0].Embedding) > 0 {
		return openAIResponse.Data[0].Embedding, nil
	}

	var flatResponse struct {
		Data []float32 `json:"data"`
	}
	if err := json.Unmarshal(raw, &flatResponse); err == nil && len(flatResponse.Data) > 0 {
		return flatResponse.Data, nil
	}

	return nil, errors.New("embedding response format is unsupported")
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature *float64  `json:"temperature,omitempty"`
	MaxTokens   *int      `json:"max_tokens,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type chatCompletionStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content json.RawMessage `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}
