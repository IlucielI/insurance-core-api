package main

import (
	"log"

	"github.com/bayuanugerah/insurance-core-api/internal/adapter/database"
	"github.com/bayuanugerah/insurance-core-api/internal/config"
	"github.com/bayuanugerah/insurance-core-api/internal/repositories"
	"github.com/bayuanugerah/insurance-core-api/internal/routes"
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
	app := routes.NewRouter(cfg, productRepository, applicationRepository, reviewCheckRepository)

	log.Printf("starting %s on port %s", cfg.AppName, cfg.HTTPPort)
	if err := app.Listen(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
