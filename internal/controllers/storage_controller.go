package controllers

import (
	"strings"

	"github.com/bayuanugerah/insurance-core-api/internal/constants"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
)

type StorageController struct {
	storageService *services.StorageService
}

func NewStorageController(storageService *services.StorageService) *StorageController {
	return &StorageController{storageService: storageService}
}

func (controller *StorageController) Presign(ctx *fiber.Ctx) error {
	if controller.storageService == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": constants.ErrStorageServiceUnavailable})
	}

	objectName := strings.TrimSpace(ctx.Query("object_name"))
	if objectName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": constants.ErrStorageObjectNameRequired})
	}

	download := strings.EqualFold(strings.TrimSpace(ctx.Query("download")), "true")
	url, err := controller.storageService.GetPresignedURL(ctx.Context(), objectName, download)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": constants.ErrStoragePresignFailed})
	}

	return ctx.JSON(fiber.Map{"data": fiber.Map{"url": url}})
}
