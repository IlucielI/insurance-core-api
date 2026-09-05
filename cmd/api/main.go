package main

import (
	"log"

	"github.com/bayuanugerah/insurance-core-api/internal/routes"
	"github.com/bayuanugerah/insurance-core-api/internal/config"
)

func main() {
	cfg := config.Load()
	app := routes.NewRouter(cfg)

	log.Printf("starting %s on port %s", cfg.AppName, cfg.HTTPPort)
	if err := app.Listen(":" + cfg.HTTPPort); err != nil {
		log.Fatal(err)
	}
}
