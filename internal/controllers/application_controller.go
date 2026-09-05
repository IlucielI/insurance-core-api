package controllers

import (
	"errors"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/dtos"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/bayuanugerah/insurance-core-api/internal/validations"
	"github.com/gofiber/fiber/v2"
)

type ApplicationController struct {
	service *services.ApplicationService
}

func NewApplicationController(service *services.ApplicationService) *ApplicationController {
	return &ApplicationController{service: service}
}

func (controller *ApplicationController) Create(ctx *fiber.Ctx) error {
	slug, err := validations.ValidateProductSlug(ctx.Params("slug"))
	if err != nil {
		return badRequest(ctx, err.Error())
	}
	var request dtos.CreateApplicationRequest
	if err := ctx.BodyParser(&request); err != nil {
		return badRequest(ctx, constants.ErrApplicationBodyInvalid)
	}
	request, err = validations.ValidateApplicationRequest(request)
	if err != nil {
		return badRequest(ctx, err.Error())
	}
	application, err := controller.service.Create(ctx.Context(), slug, request)
	if err != nil {
		return applicationError(ctx, err, constants.ErrApplicationCreateFailed)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": application,
	})
}

func (controller *ApplicationController) Get(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return badRequest(ctx, constants.ErrApplicationNotFound)
	}
	application, err := controller.service.Get(ctx.Context(), id)
	if err != nil {
		return applicationError(ctx, err, constants.ErrApplicationGetFailed)
	}
	return ctx.JSON(fiber.Map{
		"data": application,
	})
}

func badRequest(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": message,
	})
}
func applicationError(ctx *fiber.Ctx, err error, fallback string) error {
	if errors.Is(err, repositories.ErrProductNotFound) || errors.Is(err, repositories.ErrApplicationNotFound) {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if constants.IsQuoteValidationError(err) {
		return badRequest(ctx, err.Error())
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": fallback,
	})
}
