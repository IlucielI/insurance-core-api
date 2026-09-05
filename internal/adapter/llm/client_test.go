package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testAPIKey = "test-api-key"

func TestNewClient(t *testing.T) {
	client, err := NewClient(Config{BaseURL: " https://example.com ", CompletionModel: "chat", EmbeddingModel: "embed"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if client.baseURL != "https://example.com/v1" || client.completionModel != "chat" || client.embeddingModel != "embed" || client.httpClient.Timeout != 30*time.Second {
		t.Fatalf("NewClient() = %+v", client)
	}
}

func TestNewClientValidatesConfig(t *testing.T) {
	tests := []Config{
		{BaseURL: "", CompletionModel: "chat", EmbeddingModel: "embed"},
		{BaseURL: "ftp://example.com", CompletionModel: "chat", EmbeddingModel: "embed"},
		{BaseURL: "https://example.com", CompletionModel: "", EmbeddingModel: "embed"},
		{BaseURL: "https://example.com", CompletionModel: "chat", EmbeddingModel: ""},
	}
	for _, tt := range tests {
		if _, err := NewClient(tt); err == nil {
			t.Fatalf("NewClient(%+v) error = nil, want error", tt)
		}
	}
}

func TestClientCreateChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s, want /v1/chat/completions", request.URL.Path)
		}
		if request.Header.Get("Authorization") != "Bearer "+testAPIKey {
			t.Fatalf("Authorization = %q", request.Header.Get("Authorization"))
		}
		writeJSON(t, writer, map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": "hello"}}},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: " " + testAPIKey + " ", CompletionModel: "chat", EmbeddingModel: "embed"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	content, err := client.CreateChatCompletion(context.Background(), ChatCompletionInput{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("CreateChatCompletion() error = %v", err)
	}
	if content != "hello" {
		t.Fatalf("CreateChatCompletion() = %q, want hello", content)
	}
}

func TestClientCreateEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(t, writer, map[string]any{
			"data": []map[string]any{{"embedding": []float32{0.1, 0.2}}},
		})
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL + "/v1", CompletionModel: "chat", EmbeddingModel: "embed"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	embedding, err := client.CreateEmbedding(context.Background(), EmbeddingInput{Text: "hello"})
	if err != nil {
		t.Fatalf("CreateEmbedding() error = %v", err)
	}
	if len(embedding) != 2 || embedding[0] != 0.1 || embedding[1] != 0.2 {
		t.Fatalf("CreateEmbedding() = %+v, want [0.1 0.2]", embedding)
	}
}

func TestClientInputValidationAndHTTPError(t *testing.T) {
	client, err := NewClient(Config{BaseURL: "https://example.com", CompletionModel: "chat", EmbeddingModel: "embed"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if _, err := client.CreateChatCompletion(context.Background(), ChatCompletionInput{}); err == nil {
		t.Fatal("CreateChatCompletion() error = nil, want messages required")
	}
	if _, err := client.CreateEmbedding(context.Background(), EmbeddingInput{Text: " "}); err == nil {
		t.Fatal("CreateEmbedding() error = nil, want text required")
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadGateway)
		writeString(t, writer, "upstream failed")
	}))
	defer server.Close()
	client, err = NewClient(Config{BaseURL: server.URL, CompletionModel: "chat", EmbeddingModel: "embed"})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	_, err = client.CreateEmbedding(context.Background(), EmbeddingInput{Text: "hello"})
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("CreateEmbedding() error = %v, want status error", err)
	}
}

func TestParseChatCompletionResponse(t *testing.T) {
	content, err := parseChatCompletionResponse([]byte(`{"choices":[{"message":{"content":[{"text":"hello"},{"text":"world"}]}}]}`))
	if err != nil {
		t.Fatalf("parseChatCompletionResponse() error = %v", err)
	}
	if content != "hello\nworld" {
		t.Fatalf("parseChatCompletionResponse() = %q, want joined parts", content)
	}

	stream := []byte("data: not-json\ndata: {\"choices\":[{\"delta\":{\"content\":\"he\"}}]}\ndata: {\"choices\":[{\"delta\":{\"content\":\"llo\"}}]}\ndata: [DONE]")
	content, err = parseChatCompletionResponse(stream)
	if err != nil {
		t.Fatalf("parseChatCompletionResponse(stream) error = %v", err)
	}
	if content != "hello" {
		t.Fatalf("parseChatCompletionResponse(stream) = %q, want hello", content)
	}

	_, err = parseChatCompletionResponse([]byte(`{"choices":[{"message":{"content":"`))
	if err == nil || !strings.Contains(err.Error(), "failed to parse chat completion response") {
		t.Fatalf("parseChatCompletionResponse(broken) error = %v", err)
	}
}

func TestParseChatCompletionResponseErrors(t *testing.T) {
	tests := [][]byte{
		[]byte(`{"choices":[{"message":{"content":""}}]}`),
		[]byte("data: not-json\n"),
		[]byte("data: [DONE]\n"),
	}
	for _, tt := range tests {
		if _, err := parseChatCompletionResponse(tt); err == nil {
			t.Fatalf("parseChatCompletionResponse(%q) error = nil, want error", tt)
		}
	}
}

func TestParseEmbeddingResponse(t *testing.T) {
	embedding, err := parseEmbeddingResponse([]byte(`{"data":[{"embedding":[0.1,0.2]}]}`))
	if err != nil {
		t.Fatalf("parseEmbeddingResponse() error = %v", err)
	}
	if len(embedding) != 2 {
		t.Fatalf("parseEmbeddingResponse() = %+v, want 2 values", embedding)
	}

	embedding, err = parseEmbeddingResponse([]byte(`{"data":[0.1,0.2]}`))
	if err != nil {
		t.Fatalf("parseEmbeddingResponse(flat) error = %v", err)
	}
	if len(embedding) != 2 {
		t.Fatalf("parseEmbeddingResponse(flat) = %+v, want 2 values", embedding)
	}

	if _, err := parseEmbeddingResponse([]byte(`{"data":[]}`)); err == nil {
		t.Fatal("parseEmbeddingResponse(empty) error = nil, want error")
	}
}

func writeJSON(t *testing.T, writer http.ResponseWriter, value any) {
	t.Helper()
	if err := json.NewEncoder(writer).Encode(value); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
}

func writeString(t *testing.T, writer http.ResponseWriter, value string) {
	t.Helper()
	if _, err := fmt.Fprint(writer, value); err != nil {
		t.Fatalf("Fprint() error = %v", err)
	}
}
