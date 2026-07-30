package main

import (
	"log"

	"codeforge/backend/internal/api"
	"codeforge/backend/internal/config"
	"codeforge/backend/internal/database"
	"codeforge/backend/internal/llm"
	"codeforge/backend/internal/storage"
)

func main() {
	cfg := config.Load()
	services, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Seed(services.DB); err != nil {
		log.Fatal("seed: ", err)
	}
	store := storage.New(cfg)
	server := api.New(cfg, services, store, llm.New(cfg))
	log.Printf("CodeForge API listening on http://localhost:%s", cfg.Port)
	if err := server.Router().Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
