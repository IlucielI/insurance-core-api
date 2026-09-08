package controllers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/gofiber/fiber/v2"
)

type AssistantChatService interface {
	ChatWithQuote(ctx context.Context, message string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error)
}

type AssistantConversationManager interface {
	ChatWithConversation(ctx context.Context, message string, conversationID string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error)
	GetConversation(ctx context.Context, conversationID string) (dtos.AssistantConversationResponse, error)
	DeleteConversation(ctx context.Context, conversationID string) error
}

type AssistantStreamService interface {
	ChatStream(ctx context.Context, req dtos.AssistantChatRequest, onEvent func(event dtos.AssistantStreamEvent) error) error
}

type AssistantController struct {
	service AssistantChatService
}

func NewAssistantController(service AssistantChatService) *AssistantController {
	return &AssistantController{service: service}
}

func (controller *AssistantController) Chat(ctx *fiber.Ctx) error {
	if controller.service == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
	}

	var request dtos.AssistantChatRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantBodyInvalid})
	}

	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantMessageRequired})
	}
	if len(request.Message) > constants.AssistantMaxMessageSize {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantMessageTooLong})
	}
	if request.Quote != nil {
		request.ProductSlug = strings.TrimSpace(firstNonEmpty(request.ProductSlug, request.Slug))
		if request.ProductSlug == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrProductSlugRequired})
		}
		quote, err := validations.ValidateProductQuoteRequest(*request.Quote)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		request.Quote = &quote
	}

	var response dtos.AssistantChatResponse
	var err error
	if mgr, ok := controller.service.(AssistantConversationManager); ok {
		response, err = mgr.ChatWithConversation(ctx.Context(), request.Message, request.ConversationID, request.Quote, request.ProductSlug)
	} else {
		response, err = controller.service.ChatWithQuote(ctx.Context(), request.Message, request.Quote, request.ProductSlug)
	}

	if err != nil {
		if errors.Is(err, constants.ErrAssistantMessageRequiredError) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantMessageRequired})
		}
		if errors.Is(err, constants.ErrAssistantMessageTooLongError) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantMessageTooLong})
		}
		if errors.Is(err, constants.ErrAssistantServiceUnavailableError) {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
		}
		if errors.Is(err, constants.QuoteSumAssuredOutOfRangeError) || errors.Is(err, constants.QuotePaymentTermOutOfRangeError) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": constants.ErrAssistantChatFailed})
	}

	return ctx.JSON(response)
}

func (controller *AssistantController) GetConversation(ctx *fiber.Ctx) error {
	if controller.service == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
	}
	id := strings.TrimSpace(ctx.Params("id"))
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrConversationIDRequired})
	}
	mgr, ok := controller.service.(AssistantConversationManager)
	if !ok || mgr == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
	}
	conv, err := mgr.GetConversation(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrConversationNotFoundError) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": constants.ErrConversationNotFound})
		}
		if errors.Is(err, constants.ErrAssistantServiceUnavailableError) {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": constants.ErrConversationGetFailed})
	}
	return ctx.JSON(fiber.Map{"data": conv})
}

func (controller *AssistantController) DeleteConversation(ctx *fiber.Ctx) error {
	if controller.service == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
	}
	id := strings.TrimSpace(ctx.Params("id"))
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrConversationIDRequired})
	}
	mgr, ok := controller.service.(AssistantConversationManager)
	if !ok || mgr == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
	}
	err := mgr.DeleteConversation(ctx.Context(), id)
	if err != nil {
		if errors.Is(err, constants.ErrAssistantServiceUnavailableError) {
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": constants.ErrConversationDeleteFailed})
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (controller *AssistantController) ChatStream(ctx *fiber.Ctx) error {
	if controller.service == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrAssistantServiceUnavailable})
	}

	streamService, ok := controller.service.(AssistantStreamService)
	if !ok {
		return ctx.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"error": "streaming not supported"})
	}

	var request dtos.AssistantChatRequest
	if err := ctx.BodyParser(&request); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantBodyInvalid})
	}

	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantMessageRequired})
	}
	if len(request.Message) > constants.AssistantMaxMessageSize {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrAssistantMessageTooLong})
	}

	if request.Quote != nil {
		request.ProductSlug = strings.TrimSpace(firstNonEmpty(request.ProductSlug, request.Slug))
		if request.ProductSlug == "" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrProductSlugRequired})
		}
		quote, err := validations.ValidateProductQuoteRequest(*request.Quote)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		request.Quote = &quote
	}

	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")
	ctx.Set("X-Accel-Buffering", "no")

	ctx.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		streamCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err := streamService.ChatStream(streamCtx, request, func(event dtos.AssistantStreamEvent) error {
			data, err := json.Marshal(event)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return w.Flush()
		})
		if err != nil {
			log.Printf("[AssistantController] stream error: %v", err)
		}
	})

	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
