package main

import (
	"context"
	"log"
	"time"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/database"
	llmadapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/llm"
	smtpadapter "github.com/bayuanugerah/insurance-core-api/internal/adapter/smtp"
	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/ports"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
	"github.com/bayuanugerah/insurance-core-api/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	postgres, err := database.NewPostgres(database.PostgresConfig{
		DatabaseURL: cfg.DatabaseURL,
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

	productRepository := repositories.NewPostgresProductRepository(postgres.DB())
	applicationRepository := repositories.NewPostgresApplicationRepository(postgres.DB())
	reviewCheckRepository := repositories.NewPostgresApplicationReviewCheckRepository(postgres.DB())
	knowledgeRepository := repositories.NewPostgresKnowledgeRepository(postgres.DB())

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

	productService := services.NewProductService(productRepository)
	assistantService := services.NewAssistantService(knowledgeRepository, assistantLLM, productService)
	if assistantLLM != nil {
		seedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := assistantService.SeedDefaultKnowledge(seedCtx); err != nil {
			log.Printf("assistant knowledge seed failed: %v", err)
		}
	}

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

	app := routes.NewRouter(cfg, productRepository, applicationRepository, reviewCheckRepository, assistantService, mailer)

	log.Printf("starting %s on port %s", cfg.AppName, cfg.HTTPPort)
	if err := app.Listen(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
