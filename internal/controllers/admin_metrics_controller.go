package controllers

import (
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AdminMetricsController struct {
	service services.MetricsService
}

func NewAdminMetricsController(service services.MetricsService) *AdminMetricsController {
	return &AdminMetricsController{service: service}
}

func (c *AdminMetricsController) GetMetrics(ctx *fiber.Ctx) error {
	metrics, err := c.service.GetAdminMetrics(ctx.Context())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve admin metrics",
		})
	}

	return ctx.JSON(fiber.Map{
		"data": metrics,
	})
}
