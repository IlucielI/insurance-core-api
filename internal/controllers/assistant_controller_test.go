package controllers

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/gofiber/fiber/v2"
)

type assistantServiceFake struct {
	response dtos.AssistantChatResponse
	err      error
}

func (f assistantServiceFake) ChatWithQuote(context.Context, string, *dtos.ProductQuoteRequest, string) (dtos.AssistantChatResponse, error) {
	return f.response, f.err
}

func TestAssistantController(t *testing.T) {
	cases := []struct {
		name, body string
		service    interface {
			ChatWithQuote(context.Context, string, *dtos.ProductQuoteRequest, string) (dtos.AssistantChatResponse, error)
		}
		status int
	}{
		{"invalid body", "{", assistantServiceFake{}, 400},
		{"empty", `{"message":" "}`, assistantServiceFake{}, 400},
		{"invalid quote", `{"message":"hitung","slug":"secure-life-plus","quote":{"age":17}}`, assistantServiceFake{}, 400},
		{"unavailable", `{"message":"hi"}`, nil, 503},
		{"service error", `{"message":"hi"}`, assistantServiceFake{err: errors.New("upstream")}, 500},
		{"ok", `{"message":"hi"}`, assistantServiceFake{response: dtos.AssistantChatResponse{Answer: "hello"}}, 200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/", NewAssistantController(tc.service).Chat)
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			response, err := app.Test(req)
			if err != nil || response.StatusCode != tc.status {
				t.Fatalf("status=%d err=%v", response.StatusCode, err)
			}
		})
	}
}

type assistantManagerFake struct {
	assistantServiceFake
	convResponse dtos.AssistantConversationResponse
	convErr      error
	deleteErr    error
}

func (f assistantManagerFake) ChatWithConversation(ctx context.Context, message string, conversationID string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error) {
	resp, err := f.ChatWithQuote(ctx, message, quote, slug)
	resp.ConversationID = conversationID
	return resp, err
}

func (f assistantManagerFake) GetConversation(ctx context.Context, conversationID string) (dtos.AssistantConversationResponse, error) {
	return f.convResponse, f.convErr
}

func (f assistantManagerFake) DeleteConversation(ctx context.Context, conversationID string) error {
	return f.deleteErr
}

func TestAssistantControllerRejectsLongMessage(t *testing.T) {
	app := fiber.New()
	app.Post("/", NewAssistantController(assistantServiceFake{}).Chat)
	request := httptest.NewRequest("POST", "/", strings.NewReader(`{"message":"`+strings.Repeat("x", 4001)+`"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil || response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d err=%v", response.StatusCode, err)
	}
}

func TestAssistantController_Conversations(t *testing.T) {
	// 1. GetConversation - ok
	mgr := assistantManagerFake{
		convResponse: dtos.AssistantConversationResponse{ID: "conv-1", Title: "Title"},
	}
	app := fiber.New()
	controller := NewAssistantController(mgr)
	app.Get("/assistant/conversations/:id", controller.GetConversation)
	app.Delete("/assistant/conversations/:id", controller.DeleteConversation)

	req := httptest.NewRequest("GET", "/assistant/conversations/conv-1", nil)
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != fiber.StatusOK {
		t.Fatalf("GetConversation status = %d, err = %v", resp.StatusCode, err)
	}

	// 2. DeleteConversation - ok
	reqDel := httptest.NewRequest("DELETE", "/assistant/conversations/conv-1", nil)
	respDel, err := app.Test(reqDel)
	if err != nil || respDel.StatusCode != fiber.StatusNoContent {
		t.Fatalf("DeleteConversation status = %d, err = %v", respDel.StatusCode, err)
	}

	// 3. Controller with nil service - 503
	nilApp := fiber.New()
	nilController := NewAssistantController(nil)
	nilApp.Get("/assistant/conversations/:id", nilController.GetConversation)
	nilApp.Delete("/assistant/conversations/:id", nilController.DeleteConversation)

	req503 := httptest.NewRequest("GET", "/assistant/conversations/conv-1", nil)
	resp503, err := nilApp.Test(req503)
	if err != nil || resp503.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("GetConversation(nil) status = %d, want 503", resp503.StatusCode)
	}
}
