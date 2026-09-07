package controllers

import (
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type KnowledgeDocumentController struct {
	service services.KnowledgeDocumentService
}

func NewKnowledgeDocumentController(service services.KnowledgeDocumentService) *KnowledgeDocumentController {
	return &KnowledgeDocumentController{service: service}
}

func (c *KnowledgeDocumentController) List(ctx *fiber.Ctx) error {
	category := ctx.Query("category")
	status := ctx.Query("status")
	search := ctx.Query("search")

	docs, err := c.service.GetDocuments(ctx.Context(), category, status, search)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data": docs,
	})
}

func (c *KnowledgeDocumentController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocIDRequired,
		})
	}

	doc, err := c.service.GetDocumentByID(ctx.Context(), id)
	if err != nil {
		if err.Error() == constants.ErrKnowledgeDocNotFound {
			// Fallback: try by slug
			if docBySlug, errSlug := c.service.GetDocumentBySlug(ctx.Context(), id); errSlug == nil {
				return ctx.JSON(fiber.Map{
					"data": docBySlug,
				})
			}
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrKnowledgeDocNotFound,
			})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data": doc,
	})
}

func (c *KnowledgeDocumentController) GetBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	if slug == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocSlugInvalid,
		})
	}

	doc, err := c.service.GetDocumentBySlug(ctx.Context(), slug)
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
		"data": doc,
	})
}

func (c *KnowledgeDocumentController) Create(ctx *fiber.Ctx) error {
	var req dtos.CreateKnowledgeDocumentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocRequestBodyInvalid,
		})
	}

	doc, err := c.service.CreateDocument(ctx.Context(), req)
	if err != nil {
		if errors.Is(err, errors.New(constants.ErrKnowledgeDocSlugAlreadyExists)) || err.Error() == constants.ErrKnowledgeDocSlugAlreadyExists {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": doc,
	})
}

func (c *KnowledgeDocumentController) Update(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocIDRequired,
		})
	}

	var req dtos.UpdateKnowledgeDocumentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocRequestBodyInvalid,
		})
	}

	doc, err := c.service.UpdateDocument(ctx.Context(), id, req)
	if err != nil {
		if err.Error() == constants.ErrKnowledgeDocNotFound {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrKnowledgeDocNotFound,
			})
		}
		if err.Error() == constants.ErrKnowledgeDocSlugAlreadyExists {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data": doc,
	})
}

func (c *KnowledgeDocumentController) Delete(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocIDRequired,
		})
	}

	err := c.service.DeleteDocument(ctx.Context(), id)
	if err != nil {
		if err.Error() == constants.ErrKnowledgeDocNotFound {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrKnowledgeDocNotFound,
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": constants.ErrKnowledgeDocDeleteFailed,
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "knowledge document deleted successfully",
	})
}
