package controllers

import (
	"errors"
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type NotificationController struct {
	service services.NotificationService
}

func NewNotificationController(service services.NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

func (ctrl *NotificationController) List(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var query dtos.NotificationQuery
	if err := c.QueryParser(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := ctrl.service.List(ctx, query)
	if err != nil {
		if isNotificationValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":         result.Data,
		"total":        result.Total,
		"unread_count": result.UnreadCount,
	})
}

func (ctrl *NotificationController) Create(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dtos.CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	result, err := ctrl.service.Create(ctx, req)
	if err != nil {
		if isNotificationValidationError(err) {
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

func (ctrl *NotificationController) MarkAsRead(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := strings.TrimSpace(c.Params("id"))
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	result, err := ctrl.service.MarkAsRead(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotificationNotFoundError) || strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": constants.ErrNotificationNotFound,
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

func (ctrl *NotificationController) MarkAllAsRead(c *fiber.Ctx) error {
	ctx := c.UserContext()
	result, err := ctrl.service.MarkAllAsRead(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": result,
	})
}

func (ctrl *NotificationController) GetUnreadCount(c *fiber.Ctx) error {
	ctx := c.UserContext()
	count, err := ctrl.service.GetUnreadCount(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"unread_count": count,
	})
}

func isNotificationValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == constants.ErrNotificationTitleRequired ||
		msg == constants.ErrNotificationTypeRequired ||
		msg == constants.ErrNotificationMessageRequired ||
		msg == constants.ErrNotificationCategoryInvalid ||
		msg == constants.ErrNotificationSeverityInvalid ||
		msg == constants.ErrNotificationLimitInvalid
}
