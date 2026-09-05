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
