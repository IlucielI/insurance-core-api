package controllers

import (
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AdminHealthController struct {
	service services.SystemHealthService
}

func NewAdminHealthController(service services.SystemHealthService) *AdminHealthController {
	return &AdminHealthController{service: service}
}

func (ctrl *AdminHealthController) GetOverview(c *fiber.Ctx) error {
	ctx := c.UserContext()
	overview, err := ctrl.service.GetOverview(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": overview,
	})
}

func (ctrl *AdminHealthController) PingServices(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dtos.PingServicesRequest
	if err := c.BodyParser(&req); err != nil {
		// Body is optional, continue with empty service ID if body parsing encounters issue
		req.ServiceID = ""
	}

	servicesList, err := ctrl.service.PingServices(ctx, req.ServiceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": servicesList,
	})
}

func (ctrl *AdminHealthController) PingRoutes(c *fiber.Ctx) error {
	ctx := c.UserContext()
	routesList, err := ctrl.service.PingRoutes(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": routesList,
	})
}
