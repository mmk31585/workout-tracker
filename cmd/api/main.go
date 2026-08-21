package main

import (
	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		return
	}
	logger := logger.New(cfg.App.AppEnv, cfg.Log.LogLevel, cfg.Log.LogFormat)
	logger.Info("Server initilais complited :)")
}
