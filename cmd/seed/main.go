package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/logger"
	"github.com/mmk31585/workout-tracker/internal/seed"
	"github.com/mmk31585/workout-tracker/internal/storage"
)

func main() {
	log.Print("Start seeding...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logger.New(cfg.App.AppEnv, cfg.Log.LogLevel, cfg.Log.LogFormat)

	logger.Info("Starting seeder...")

	store, err := storage.New(cfg.DB)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	repo := seed.NewPostgresSeedRepository(store.DB())
	svc := seed.NewSeederService(repo, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := svc.Run(ctx); err != nil {
		logger.Error("seeding failed", "error", err)
		os.Exit(1)
	}
	logger.Info("seeding completed successfully")
}
