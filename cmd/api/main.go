package main

import (
	"context"
	"fmt"
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
	"github.com/mmk31585/workout-tracker/internal/metrics"
	"github.com/mmk31585/workout-tracker/internal/server"
	servermiddleware "github.com/mmk31585/workout-tracker/internal/server/server_middleware"
	"github.com/mmk31585/workout-tracker/internal/storage"
	"github.com/mmk31585/workout-tracker/internal/user"
	workoutplan "github.com/mmk31585/workout-tracker/internal/workout_plan"
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

	workoutPlanRepo := workoutplan.NewWorkoutPlanRepository(store.DB())
	workoutItemRepo := workoutplan.NewWorkoutPlanItemRepository(store.DB())
	workoutPlanService := workoutplan.NewWorkoutPlanService(workoutPlanRepo, workoutItemRepo, store.DB())
	workoutPlanHandler := workoutplan.NewWorkoutPlanHandler(workoutPlanService)

	metricsInstance := metrics.NewMetrics(
		metrics.WithRuntimeStats(),
		metrics.WithStartTime(time.Now()),
	)
	authHandler.SetMetrics(metricsInstance)
	workoutPlanHandler.SetMetrics(metricsInstance)

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
			WorkoutPlan: server.WorkoutPlanHandlers{
				CreatePlan: workoutPlanHandler.CreatePlan,
				GetPlan:    workoutPlanHandler.GetPlan,
				ListPlans:  workoutPlanHandler.ListPlans,
				UpdatePlan: workoutPlanHandler.UpdatePlan,
				DeletePlan: workoutPlanHandler.DeletePlan,
			},
		},
		AuthMiddleware:      authMiddleware,
		BasicAuthMiddleware: basicAuthMiddleware,
		Storage:             store,
	}
	handler := server.RegisterRouter(routerCfg)

	fmt.Println("DEBUG: Router registered successfully")
	srv := &http.Server{
		Handler:           handler,
		Addr:              net.JoinHostPort(cfg.Http.HttpHost, cfg.Http.HttpPort),
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Println("DEBUG: HTTP server created, starting goroutine")
	serverErr := make(chan error, 1)

	go func() {
		fmt.Println("DEBUG: Inside server goroutine, about to log listening")
		logger.Info("server listening", "address", srv.Addr)

		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			fmt.Println("DEBUG: ListenAndServe error:", err)
			serverErr <- err
		}
		fmt.Println("DEBUG: ListenAndServe returned")
	}()
	fmt.Println("DEBUG: Goroutine started, entering select")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-serverErr:
		logger.Error("server failed", "error", err)
		os.Exit(1)

	case sig := <-quit:
		logger.Info("shutdown signal received", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}
