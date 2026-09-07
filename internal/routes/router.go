package routes

import (
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/controllers"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewRouter(cfg config.Config, productRepository repositories.ProductRepository, applicationRepository repositories.ApplicationRepository, reviewCheckRepository repositories.ApplicationReviewCheckRepository, assistantService *services.AssistantService, storageService *services.StorageService, mailer ports.Mailer, messageBus ports.MessageBus, optionalRepositories ...any) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	var questionnaireRepository repositories.QuestionnaireRepository
	var pricingRuleRepository repositories.PricingRuleRepository

	for _, repo := range optionalRepositories {
		switch r := repo.(type) {
		case repositories.QuestionnaireRepository:
			questionnaireRepository = r
		case repositories.PricingRuleRepository:
			pricingRuleRepository = r
		}
	}

	healthController := controllers.NewHealthController(cfg.Version, cfg.GitHash, time.Now())
	productService := services.NewProductService(productRepository, pricingRuleRepository)
	productController := controllers.NewProductController(productService)
	questionnaireService := services.NewQuestionnaireService(questionnaireRepository, productRepository)
	questionnaireController := controllers.NewQuestionnaireController(questionnaireService)
	applicationService := services.NewApplicationService(productRepository, applicationRepository, reviewCheckRepository, productService, mailer, messageBus, questionnaireRepository)
	applicationController := controllers.NewApplicationController(applicationService)
	assistantController := controllers.NewAssistantController(assistantService)
	storageController := controllers.NewStorageController(storageService)

	app.Get("/health", healthController.Check)

	api := app.Group("/api/v1")
	api.Get("/products", productController.List)
	api.Get("/products/:slug", productController.Detail)
	api.Get("/products/:slug/pricing-rules", productController.GetPricingRules)
	api.Post("/products/:slug/quotes", productController.CreateQuote)
	api.Get("/products/:slug/questionnaire", questionnaireController.GetByProduct)
	api.Get("/questionnaires", questionnaireController.GetDefault)
	api.Post("/products/:slug/applications", applicationController.Create)
	api.Post("/applications", applicationController.Create)
	api.Get("/applications", applicationController.List)
	api.Get("/applications/:id", applicationController.Get)
	api.Patch("/applications/:id/status", applicationController.UpdateStatus)
	api.Get("/applications/:id/review-checks", applicationController.ListReviewChecks)
	api.Patch("/applications/:id/review-checks/:check_type", applicationController.UpdateReviewCheck)
	api.Post("/assistant/chat", assistantController.Chat)
	api.Get("/storage/presign", storageController.Presign)

	return app
}
