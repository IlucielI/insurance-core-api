package controllers

import (
	"context"
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/gofiber/fiber/v2"
)

type AssistantChatService interface {
	ChatWithQuote(ctx context.Context, message string, quote *dtos.ProductQuoteRequest, slug string) (dtos.AssistantChatResponse, error)
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

	response, err := controller.service.ChatWithQuote(ctx.Context(), request.Message, request.Quote, request.ProductSlug)
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
