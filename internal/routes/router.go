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
	var metricsRepository repositories.MetricsRepository
	var metricsService services.MetricsService

	for _, repo := range optionalRepositories {
		switch r := repo.(type) {
		case repositories.QuestionnaireRepository:
			questionnaireRepository = r
		case repositories.PricingRuleRepository:
			pricingRuleRepository = r
		case repositories.MetricsRepository:
			metricsRepository = r
		case services.MetricsService:
			metricsService = r
		}
	}

	if metricsService == nil && metricsRepository != nil {
		metricsService = services.NewMetricsService(metricsRepository)
	}

	healthController := controllers.NewHealthController(cfg.Version, cfg.GitHash, time.Now())
	productService := services.NewProductService(productRepository, pricingRuleRepository)
	productController := controllers.NewProductController(productService)
	questionnaireService := services.NewQuestionnaireService(questionnaireRepository, productRepository)
	questionnaireController := controllers.NewQuestionnaireController(questionnaireService)
	applicationService := services.NewApplicationService(productRepository, applicationRepository, reviewCheckRepository, productService, mailer, messageBus, questionnaireRepository)
	if applicationService != nil && cfg.CustomerAppBaseURL != "" {
		applicationService.SetAppBaseURL(cfg.CustomerAppBaseURL)
	}
	applicationController := controllers.NewApplicationController(applicationService)
	var assistantChatService controllers.AssistantChatService
	if assistantService != nil {
		if applicationService != nil {
			assistantService.WithApplicationService(applicationService)
		}
		assistantChatService = assistantService
	}
	assistantController := controllers.NewAssistantController(assistantChatService)
	storageController := controllers.NewStorageController(storageService)

	app.Get("/health", healthController.Check)

	api := app.Group("/api/v1")
	api.Get("/products", productController.List)
	api.Post("/products", productController.Create)
	api.Get("/products/id/:id", productController.GetByID)
	api.Put("/products/:id", productController.Update)
	api.Patch("/products/:id/status", productController.UpdateStatus)
	api.Post("/products/:id/toggle-status", productController.ToggleStatus)
	api.Delete("/products/:id", productController.Delete)
	api.Get("/products/:slug", productController.Detail)
	api.Get("/products/:slug/pricing-rules", productController.GetPricingRules)
	api.Put("/products/:slug/pricing-rules", productController.UpdatePricingRules)
	api.Post("/products/:slug/quotes", productController.CreateQuote)
	api.Get("/admin/products/metrics", productController.GetMetrics)
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
	api.Post("/assistant/chat/stream", assistantController.ChatStream)
	api.Get("/assistant/conversations/:id", assistantController.GetConversation)
	api.Delete("/assistant/conversations/:id", assistantController.DeleteConversation)
	api.Get("/storage/presign", storageController.Presign)
	if metricsService != nil {
		adminMetricsController := controllers.NewAdminMetricsController(metricsService)
		api.Get("/admin/metrics", adminMetricsController.GetMetrics)
	}

	return app
}
