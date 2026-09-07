package controllers

import (
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AuditLogController struct {
	service services.AuditLogService
}

func NewAuditLogController(service services.AuditLogService) *AuditLogController {
	return &AuditLogController{service: service}
}

func (ctrl *AuditLogController) List(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var query dtos.AuditLogQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := ctrl.service.List(ctx, query)
	if err != nil {
		if isAuditValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":  result.Data,
		"total": result.Total,
	})
}

func (ctrl *AuditLogController) GetByID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	result, err := ctrl.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrAuditLogNotFoundError) || strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrAuditLogNotFound,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": result,
	})
}

func (ctrl *AuditLogController) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dtos.CreateAuditLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := ctrl.service.Record(ctx, req)
	if err != nil {
		if isAuditValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": result,
	})
}

func isAuditValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == constants.ErrAuditLogActorNameRequired ||
		msg == constants.ErrAuditLogActorRoleRequired ||
		msg == constants.ErrAuditLogActionRequired ||
		msg == constants.ErrAuditLogCategoryInvalid ||
		msg == constants.ErrAuditLogStatusInvalid ||
		msg == constants.ErrAuditLogTargetRequired ||
		msg == constants.ErrAuditLogQueryLimitInvalid
}
