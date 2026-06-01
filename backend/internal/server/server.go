// Package server конфигурирует HTTP сервер и роутинг.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/config"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

// Server инкапсулирует зависимости HTTP сервера.
type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *sqlx.DB
	redis  *redis.Client
	router *chi.Mux
	server *http.Server
}

// New создаёт сервер с подключёнными middleware и роутами.
func New(cfg *config.Config, log *slog.Logger, db *sqlx.DB, redisClient *redis.Client) *Server {
	s := &Server{
		cfg:   cfg,
		log:   log,
		db:    db,
		redis: redisClient,
	}

	s.setupRouter()

	s.server = &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           s.router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return s
}

func (s *Server) setupRouter() {
	r := chi.NewRouter()

	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(middleware.Recovery(s.log))
	r.Use(middleware.Logging(s.log))
	r.Use(chimw.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/version", s.handleVersion)

		// Здесь будут регистрироваться feature-модули в следующих этапах.
		// Каждая фича добавляет свои роуты через RegisterRoutes(r).
	})

	s.router = r
}

// handleHealth — проверка живости сервиса. Используется для readiness/liveness проб.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{
		"server":   "ok",
		"database": "ok",
		"redis":    "ok",
	}

	if err := s.db.PingContext(r.Context()); err != nil {
		checks["database"] = "fail"
	}
	if err := s.redis.Ping(r.Context()).Err(); err != nil {
		checks["redis"] = "fail"
	}

	response.OK(w, map[string]any{
		"status": "ok",
		"checks": checks,
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]any{
		"version": s.cfg.Server.Version,
		"env":     s.cfg.Server.Env,
	})
}

// Start запускает HTTP сервер. Блокирующий вызов.
func (s *Server) Start() error {
	s.log.Info("starting http server",
		slog.String("port", s.cfg.Server.Port),
		slog.String("env", s.cfg.Server.Env),
	)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown корректно останавливает сервер с таймаутом.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down http server")
	return s.server.Shutdown(ctx)
}
