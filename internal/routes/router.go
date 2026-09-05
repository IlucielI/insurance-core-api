package routes

import (
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/controllers"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewRouter(cfg config.Config, productRepository repositories.ProductRepository) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	healthController := controllers.NewHealthController(cfg.Version, cfg.GitHash, time.Now())
	productService := services.NewProductService(productRepository)
	productController := controllers.NewProductController(productService)

	app.Get("/health", healthController.Check)

	api := app.Group("/api/v1")
	api.Get("/products", productController.List)
	api.Get("/products/:slug", productController.Detail)

	return app
}
