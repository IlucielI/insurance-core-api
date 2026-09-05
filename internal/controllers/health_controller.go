package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthController struct {
	version   string
	gitHash   string
	startedAt time.Time
}

func NewHealthController(version string, gitHash string, startedAt time.Time) *HealthController {
	return &HealthController{
		version:   version,
		gitHash:   gitHash,
		startedAt: startedAt,
	}
}

func (controller *HealthController) Check(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"version":  controller.version,
		"uptime":   time.Since(controller.startedAt).String(),
		"git_hash": controller.gitHash,
	})
}
