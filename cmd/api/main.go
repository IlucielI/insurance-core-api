package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/database"
	llmadapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	natsadapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/nats"
	redisadapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/redis"
	s3adapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/s3"
	smtpadapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/smtp"
	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
	emailtemplate "github.com/bayuanugerah/insurance-core-api/internal/templates/email"
	"github.com/bayuanugerah/insurance-core-api/internal/workers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	postgres, err := database.NewPostgres(database.PostgresConfig{
		DatabaseURL:     cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: time.Duration(cfg.DBConnMaxLifetimeMin) * time.Minute,
		ConnMaxIdleTime: time.Duration(cfg.DBConnMaxIdleTimeMin) * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := postgres.Close(); err != nil {
			log.Printf("failed to close database connection: %v", err)
		}
	}()

	if err := database.RunMigrations(postgres.DB()); err != nil {
		log.Fatal(err)
	}

	var redisClient *redisadapter.Client
	if cfg.RedisHost != "" {
		client, err := redisadapter.NewClient(redisadapter.Config{
			Host:     cfg.RedisHost,
			Port:     cfg.RedisPort,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
			// cfg.RedisTimeout is configured in seconds (integer)
			Timeout: time.Duration(cfg.RedisTimeout) * time.Second,
		})
		if err != nil {
			log.Printf("redis disabled: %v", err)
		} else {
			redisClient = client
			defer func(c *redisadapter.Client) {
				if c != nil {
					if err := c.Close(); err != nil {
						log.Printf("failed to close redis connection: %v", err)
					}
				}
			}(client)
		}
	}

	productRepository := repositories.NewPostgresProductRepository(postgres.DB())
	if redisClient != nil {
		productRepository = productRepository.WithCache(redisClient)
	}
	applicationRepository := repositories.NewPostgresApplicationRepository(postgres.DB())
	reviewCheckRepository := repositories.NewPostgresApplicationReviewCheckRepository(postgres.DB())
	questionnaireRepository := repositories.NewPostgresQuestionnaireRepository(postgres.DB(), redisClient)
	pricingRuleRepository := repositories.NewPostgresPricingRuleRepository(postgres.DB(), redisClient)
	knowledgeRepository := repositories.NewPostgresKnowledgeRepository(postgres.DB())
	knowledgeDocRepository := repositories.NewPostgresKnowledgeDocumentRepository(postgres.DB())
	assistantConversationRepository := repositories.NewPostgresAssistantConversationRepository(postgres.DB(), redisClient)
	metricsRepository := repositories.NewPostgresMetricsRepository(postgres.DB())
	knowledgeMetricsRepository := repositories.NewPostgresKnowledgeMetricsRepository(postgres.DB())
	auditLogRepository := repositories.NewPostgresAuditLogRepository(postgres.DB())
	auditLogService := services.NewAuditLogService(auditLogRepository)
	systemHealthService := services.NewSystemHealthService(postgres.DB(), redisClient, auditLogRepository, cfg, time.Now().UTC())
	notificationRepository := repositories.NewPostgresNotificationRepository(postgres.DB())
	notificationService := services.NewNotificationService(notificationRepository)

	var assistantLLM services.AssistantLLM
	if cfg.LLMBaseURL != "" && cfg.LLMCompletionModel != "" && cfg.LLMEmbeddingModel != "" {
		llmClient, err := llmadapter.NewClient(llmadapter.Config{
			BaseURL:         cfg.LLMBaseURL,
			APIKey:          cfg.LLMAPIKey,
			CompletionModel: cfg.LLMCompletionModel,
			EmbeddingModel:  cfg.LLMEmbeddingModel,
		})
		if err != nil {
			log.Printf("assistant disabled: %v", err)
		} else {
			assistantLLM = llmClient
		}
	}

	productService := services.NewProductService(productRepository, pricingRuleRepository)
	assistantService := services.NewAssistantService(knowledgeRepository, assistantLLM, productService, assistantConversationRepository)
	knowledgeRAGService := services.NewKnowledgeRAGService(knowledgeDocRepository, knowledgeRepository, knowledgeMetricsRepository, assistantLLM)
	if assistantLLM != nil {
		seedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := assistantService.SeedDefaultKnowledge(seedCtx); err != nil {
			log.Printf("assistant knowledge seed failed: %v", err)
		}
	}

	var natsClient *natsadapter.Client
	if cfg.NATSHost != "" {
		client, err := natsadapter.NewClient(natsadapter.Config{
			Host:    cfg.NATSHost,
			Port:    cfg.NATSPort,
			Token:   cfg.NATSToken,
			Name:    cfg.NATSName,
			Timeout: time.Duration(cfg.NATSTimeout) * time.Second,
		})
		if err != nil {
			log.Printf("nats disabled: %v", err)
		} else {
			natsClient = client
			defer natsClient.Close()
		}
	}

	storageRepository := initStorageRepository(cfg)
	storageService := services.NewStorageService(storageRepository)

	var mailer ports.Mailer
	if cfg.SMTPHost != "" && cfg.SMTPFromEmail != "" {
		smtpClient, err := smtpadapter.NewClient(smtpadapter.Config{
			Host:       cfg.SMTPHost,
			Port:       cfg.SMTPPort,
			Username:   cfg.SMTPUsername,
			Password:   cfg.SMTPPassword,
			FromEmail:  cfg.SMTPFromEmail,
			FromName:   cfg.SMTPFromName,
			Encryption: smtpadapter.Encryption(cfg.SMTPEncryption),
		})
		if err != nil {
			log.Printf("smtp disabled: %v", err)
		} else {
			mailer = smtpClient
		}
	}

	if natsClient != nil && mailer != nil {
		renderer, err := emailtemplate.NewRenderer()
		if err != nil {
			log.Printf("failed to initialize email template renderer: %v", err)
		} else {
			emailWorker := workers.NewEmailNotificationWorker(natsClient, mailer, renderer, cfg.CustomerAppBaseURL)
			if err := emailWorker.Start(context.Background()); err != nil {
				log.Printf("failed to start email worker: %v", err)
			} else {
				defer emailWorker.Stop()
			}
		}
	}

	if natsClient != nil && notificationService != nil {
		adminNotifWorker := workers.NewAdminNotificationWorker(natsClient, notificationService, applicationRepository)
		if err := adminNotifWorker.Start(context.Background()); err != nil {
			log.Printf("failed to start admin notification worker: %v", err)
		} else {
			defer adminNotifWorker.Stop()
		}
	}

	app := routes.NewRouter(cfg, productRepository, applicationRepository, reviewCheckRepository, assistantService, storageService, mailer, natsClient, questionnaireRepository, pricingRuleRepository, metricsRepository, knowledgeDocRepository, knowledgeRepository, knowledgeMetricsRepository, knowledgeRAGService, auditLogRepository, auditLogService, systemHealthService, notificationRepository, notificationService)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("starting %s on port %s", cfg.AppName, cfg.HTTPPort)
		if err := app.Listen(":" + cfg.HTTPPort); err != nil {
			serverErrChan <- err
		}
	}()

	select {
	case sig := <-shutdownChan:
		log.Printf("received signal %v, gracefully shutting down %s...", sig, cfg.AppName)
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("error during server shutdown: %v", err)
		}
	case err := <-serverErrChan:
		log.Printf("server listener stopped: %v", err)
	}
}

func initStorageRepository(cfg config.Config) repositories.StorageRepository {
	if cfg.S3Endpoint == "" || cfg.S3AccessKey == "" || cfg.S3SecretKey == "" || cfg.S3Bucket == "" {
		return nil
	}

	s3Client, err := s3adapter.NewClient(s3adapter.Config{
		Endpoint:       cfg.S3Endpoint,
		AccessKey:      cfg.S3AccessKey,
		SecretKey:      cfg.S3SecretKey,
		Region:         cfg.S3Region,
		UseSSL:         cfg.S3UseSSL,
		ForcePathStyle: cfg.S3ForcePathStyle,
		PresignExpiry:  time.Duration(cfg.S3UploadUrlLifetime) * time.Minute,
	})
	if err != nil {
		log.Printf("s3 disabled: %v", err)
		return nil
	}

	bootstrapCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s3Client.EnsureBucketExists(bootstrapCtx, cfg.S3Bucket); err != nil {
		log.Printf("s3 bucket bootstrap failed: %v", err)
		return nil
	}

	repository, err := repositories.NewS3StorageRepository(
		s3Client,
		cfg.S3Bucket,
		time.Duration(cfg.S3UploadUrlLifetime)*time.Minute,
		time.Duration(cfg.S3DownloadUrlLifetime)*time.Minute,
		cfg.S3OverrideBaseURL,
	)
	if err != nil {
		log.Printf("s3 repository disabled: %v", err)
		return nil
	}

	return repository
}
