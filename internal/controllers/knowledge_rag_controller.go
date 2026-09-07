package controllers

import (
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type KnowledgeRAGController struct {
	service services.KnowledgeRAGService
}

func NewKnowledgeRAGController(service services.KnowledgeRAGService) *KnowledgeRAGController {
	return &KnowledgeRAGController{service: service}
}

func (c *KnowledgeRAGController) Reindex(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if strings.TrimSpace(id) == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocIDRequired,
		})
	}

	resp, err := c.service.ReindexDocument(ctx.Context(), id)
	if err != nil {
		if err.Error() == constants.ErrKnowledgeDocNotFound {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrKnowledgeDocNotFound,
			})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data":    resp,
		"message": "document reindexed successfully",
	})
}

func (c *KnowledgeRAGController) GetMetrics(ctx *fiber.Ctx) error {
	metrics, err := c.service.GetMetrics(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve knowledge base metrics",
		})
	}

	return ctx.JSON(fiber.Map{
		"data": metrics,
	})
}

func (c *KnowledgeRAGController) SimulateChat(ctx *fiber.Ctx) error {
	var req dtos.SimulateChatRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if strings.TrimSpace(req.Query) == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query is required",
		})
	}

	resp, err := c.service.SimulateChat(ctx.Context(), req)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data": resp,
	})
}
