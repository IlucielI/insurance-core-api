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
	var knowledgeDocRepository repositories.KnowledgeDocumentRepository
	var knowledgeDocService services.KnowledgeDocumentService
	var knowledgeRepository repositories.KnowledgeRepository
	var knowledgeMetricsRepository repositories.KnowledgeMetricsRepository
	var knowledgeRAGService services.KnowledgeRAGService
	var auditLogRepository repositories.AuditLogRepository
	var auditLogService services.AuditLogService
	var systemHealthService services.SystemHealthService
	var notificationRepository repositories.NotificationRepository
	var notificationService services.NotificationService

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
		case repositories.KnowledgeDocumentRepository:
			knowledgeDocRepository = r
		case services.KnowledgeDocumentService:
			knowledgeDocService = r
		case repositories.KnowledgeRepository:
			knowledgeRepository = r
		case repositories.KnowledgeMetricsRepository:
			knowledgeMetricsRepository = r
		case services.KnowledgeRAGService:
			knowledgeRAGService = r
		case repositories.AuditLogRepository:
			auditLogRepository = r
		case services.AuditLogService:
			auditLogService = r
		case services.SystemHealthService:
			systemHealthService = r
		case repositories.NotificationRepository:
			notificationRepository = r
		case services.NotificationService:
			notificationService = r
		}
	}

	if metricsService == nil && metricsRepository != nil {
		metricsService = services.NewMetricsService(metricsRepository)
	}

	if knowledgeDocService == nil && knowledgeDocRepository != nil {
		knowledgeDocService = services.NewKnowledgeDocumentService(knowledgeDocRepository)
	}

	if knowledgeRAGService == nil && knowledgeDocRepository != nil && knowledgeRepository != nil && knowledgeMetricsRepository != nil {
		knowledgeRAGService = services.NewKnowledgeRAGService(knowledgeDocRepository, knowledgeRepository, knowledgeMetricsRepository, nil)
	}

	if auditLogService == nil && auditLogRepository != nil {
		auditLogService = services.NewAuditLogService(auditLogRepository)
	}

	if notificationService == nil && notificationRepository != nil {
		notificationService = services.NewNotificationService(notificationRepository)
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
	api.Post("/applications/:id/request-documents", applicationController.RequestDocuments)
	api.Post("/applications/:id/rfi", applicationController.RequestDocuments)
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
	if knowledgeDocService != nil {
		knowledgeDocController := controllers.NewKnowledgeDocumentController(knowledgeDocService)
		api.Get("/knowledge/documents", knowledgeDocController.List)
		api.Post("/knowledge/documents", knowledgeDocController.Create)
		api.Get("/knowledge/documents/slug/:slug", knowledgeDocController.GetBySlug)
		api.Get("/knowledge/documents/:id", knowledgeDocController.Get)
		api.Put("/knowledge/documents/:id", knowledgeDocController.Update)
		api.Delete("/knowledge/documents/:id", knowledgeDocController.Delete)
	}

	if knowledgeRAGService != nil {
		knowledgeRAGController := controllers.NewKnowledgeRAGController(knowledgeRAGService)
		api.Post("/knowledge/documents/:id/reindex", knowledgeRAGController.Reindex)
		api.Get("/admin/knowledge/metrics", knowledgeRAGController.GetMetrics)
		api.Post("/knowledge/simulate-chat", knowledgeRAGController.SimulateChat)
	}

	if systemHealthService != nil {
		adminHealthController := controllers.NewAdminHealthController(systemHealthService)
		api.Get("/admin/health/overview", adminHealthController.GetOverview)
		api.Post("/admin/health/ping", adminHealthController.PingServices)
		api.Post("/admin/health/routes/ping", adminHealthController.PingRoutes)
	}

	if auditLogService != nil {
		auditLogController := controllers.NewAuditLogController(auditLogService)
		api.Get("/admin/audit-logs", auditLogController.List)
		api.Get("/admin/audit-logs/:id", auditLogController.GetByID)
		api.Post("/admin/audit-logs", auditLogController.Create)
	}

	if notificationService != nil {
		notificationController := controllers.NewNotificationController(notificationService)
		api.Get("/admin/notifications", notificationController.List)
		api.Post("/admin/notifications", notificationController.Create)
		api.Patch("/admin/notifications/:id/read", notificationController.MarkAsRead)
		api.Post("/admin/notifications/mark-all-read", notificationController.MarkAllAsRead)
		api.Get("/admin/notifications/unread-count", notificationController.GetUnreadCount)
	}

	return app
}
