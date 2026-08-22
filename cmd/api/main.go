package main

import (
	"log"
	"net"
	"os"

	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/logger"
	"github.com/mmk31585/workout-tracker/internal/storage"
)

func main() {
	log.Print("Starting Server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logger.New(cfg.App.AppEnv, cfg.Log.LogLevel, cfg.Log.LogFormat)
	logger.Info("Server initilais complited :)")
	logger.Info("starting server", "address", net.JoinHostPort(cfg.Http.HttpHost, cfg.Http.HttpPort), "env", cfg.App.AppEnv)

	store, err := storage.New(cfg.DB)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()
}
