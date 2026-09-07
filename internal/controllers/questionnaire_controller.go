package controllers

import (
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type QuestionnaireController struct {
	service *services.QuestionnaireService
}

func NewQuestionnaireController(service *services.QuestionnaireService) *QuestionnaireController {
	return &QuestionnaireController{service: service}
}

func (c *QuestionnaireController) GetByProduct(ctx *fiber.Ctx) error {
	slug := strings.TrimSpace(ctx.Params("slug"))
	if slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product slug is required",
		})
	}

	result, err := c.service.GetByProductSlug(ctx.Context(), slug)
	if err != nil {
		if errors.Is(err, repositories.ErrProductNotFound) || errors.Is(err, repositories.ErrQuestionnaireNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "questionnaire not found for product",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve questionnaire",
		})
	}

	return ctx.JSON(fiber.Map{
		"data": result,
	})
}

func (c *QuestionnaireController) GetDefault(ctx *fiber.Ctx) error {
	// If query param product_slug or slug is passed, route to product slug
	if slug := strings.TrimSpace(ctx.Query("product_slug")); slug != "" {
		result, err := c.service.GetByProductSlug(ctx.Context(), slug)
		if err == nil {
			return ctx.JSON(fiber.Map{"data": result})
		}
	}

	result, err := c.service.GetDefault(ctx.Context())
	if err != nil {
		if errors.Is(err, repositories.ErrQuestionnaireNotFound) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "default questionnaire not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve questionnaire",
		})
	}

	return ctx.JSON(fiber.Map{
		"data": result,
	})
}
