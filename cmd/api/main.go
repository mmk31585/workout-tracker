package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mmk31585/workout-tracker/internal/auth"
	"github.com/mmk31585/workout-tracker/internal/config"
	"github.com/mmk31585/workout-tracker/internal/health"
	"github.com/mmk31585/workout-tracker/internal/logger"
	"github.com/mmk31585/workout-tracker/internal/server"
	servermiddleware "github.com/mmk31585/workout-tracker/internal/server/server_middleware"
	"github.com/mmk31585/workout-tracker/internal/storage"
	"github.com/mmk31585/workout-tracker/internal/user"
)

func main() {
	log.Print("Starting Server...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logger.New(cfg.App.AppEnv, cfg.Log.LogLevel, cfg.Log.LogFormat)
	logger.Info("Server initialized")
	logger.Info("starting server", "address", net.JoinHostPort(cfg.Http.HttpHost, cfg.Http.HttpPort), "env", cfg.App.AppEnv)

	store, err := storage.New(cfg.DB)
	if err != nil {
		logger.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	userRepo := user.NewUserRepository(store.DB())

	jwtExp, err := time.ParseDuration(cfg.Auth.JWT.JWTExp)
	if err != nil {
		logger.Error("failed to parse JWT expiration", "error", err)
		os.Exit(1)
	}

	jwtService := auth.NewJWTService(
		[]byte(cfg.Auth.JWT.JWTSecret),
		cfg.App.AppName,
		jwtExp,
	)

	authService := auth.NewAuthService(userRepo, jwtService, jwtExp)
	authHandler := auth.NewAuthHandler(authService)
	basicAuthMiddleware := servermiddleware.NewBasicAuthMiddleware(
		cfg.Auth.Basic.UserName,
		cfg.Auth.Basic.Password,
	)
	authMiddleware := servermiddleware.NewAuthMiddlware(jwtService)

	routerCfg := server.RouterConfig{
		Handlers: server.Handlers{
			Auth: server.AuthHandlers{
				Signup: authHandler.Signup,
				Login:  authHandler.Login,
				Logout: authHandler.Logout,
			},
			Health: server.HealthHandlers{
				Health: health.Health,
				Live:   health.Live,
				Ready:  health.Ready,
			},
		},
		AuthMiddleware: authMiddleware,
		BasicAuthMiddleware: basicAuthMiddleware,
		Storage:        store,
	}
	handler := server.RegisterRouter(routerCfg)

	srv := &http.Server{
		Handler:           handler,
		Addr:              net.JoinHostPort(cfg.Http.HttpHost, cfg.Http.HttpPort),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server listening", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}
