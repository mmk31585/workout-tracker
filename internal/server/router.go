package server

import (
	"expvar"
	"net/http"
	"time"

	"github.com/arl/statsviz"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/mmk31585/workout-tracker/internal/config"
	servermiddleware "github.com/mmk31585/workout-tracker/internal/server/server_middleware"
	"github.com/mmk31585/workout-tracker/internal/storage"
)

var (
	versionVar = expvar.NewString("version")
)

func init() {
	versionVar.Set(config.Version)
}

type Handlers struct {
	Auth   AuthHandlers
	Health HealthHandlers
}

type AuthHandlers struct {
	Signup http.HandlerFunc
	Login  http.HandlerFunc
	Logout http.HandlerFunc
}

type HealthHandlers struct {
	Health http.HandlerFunc
	Live   http.HandlerFunc
	Ready  http.HandlerFunc
}

type RouterConfig struct {
	Handlers            Handlers
	AuthMiddleware      *servermiddleware.AuthMiddleware
	BasicAuthMiddleware *servermiddleware.BasicAuthMiddleware
	Storage             *storage.Storage
}

func RegisterRouter(cfg RouterConfig) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RequestID,
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: false,
			MaxAge:           300,
		}),
		middleware.ClientIPFromXFF(),
		middleware.Recoverer,
		middleware.Logger,
		middleware.Timeout(60*time.Second),
	)

	statsvizSrv, err := statsviz.NewServer()
	if err != nil {
		router.Group(func(r chi.Router) {
			r.Use(cfg.BasicAuthMiddleware.BasicAuthMiddleware)
			r.Get("/debug/vars", expvar.Handler().ServeHTTP)
		})
	}

	router.Group(func(r chi.Router) {
		r.Use(cfg.BasicAuthMiddleware.BasicAuthMiddleware)

		r.Get("/debug/vars", expvar.Handler().ServeHTTP)
		r.Get("/debug/statsviz/", statsvizSrv.Index())
		r.Get("/debug/statsviz/*", statsvizSrv.Index())
		r.Get("/debug/statsviz/ws", statsvizSrv.Ws())
	})
	router.Route("/health", func(r chi.Router) {
		r.Use(cfg.BasicAuthMiddleware.BasicAuthMiddleware)
		r.Get("/", cfg.Handlers.Health.Health)
		r.Get("/live", cfg.Handlers.Health.Live)
		r.Get("/ready", cfg.Handlers.Health.Ready)
	})
	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", cfg.Handlers.Auth.Signup)
			r.Post("/login", cfg.Handlers.Auth.Login)
			r.Group(func(r chi.Router) {
				r.Use(cfg.AuthMiddleware.AuthTokenMiddleware)
				r.Post("/logout", cfg.Handlers.Auth.Logout)
			})
		})
	})
	return router
}
