// Package server конфигурирует HTTP сервер, роутинг и инициализирует зависимости.
package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	redisclient "github.com/redis/go-redis/v9"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/config"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/assignments"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/auth"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/feed"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/grades"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/groups"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/quizzes"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/minio"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/postgres"
	redisinfra "github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/redis"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/sms"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/jwt"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

// Server инкапсулирует зависимости HTTP сервера.
type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *sqlx.DB
	redis  *redisclient.Client
	router *chi.Mux
	server *http.Server
}

// New создаёт сервер с подключёнными middleware и роутами.
func New(cfg *config.Config, log *slog.Logger, db *sqlx.DB, rc *redisclient.Client) *Server {
	s := &Server{cfg: cfg, log: log, db: db, redis: rc}
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

	allowedOrigins := s.cfg.Server.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// --- Инициализация зависимостей ---
	jwtManager := jwt.NewManager(
		s.cfg.JWT.AccessSecret,
		s.cfg.JWT.RefreshSecret,
		s.cfg.JWT.AccessTTL,
		s.cfg.JWT.RefreshTTL,
	)

	var smsSender sms.Sender
	if s.cfg.SMS.Provider == "eskiz" {
		smsSender = sms.NewEskizSender(s.cfg.SMS.EskizEmail, s.cfg.SMS.EskizPassword, s.cfg.SMS.EskizFrom)
	} else {
		smsSender = sms.NewMockSender()
	}

	rateLimiter := redisinfra.NewRateLimiter(s.redis)
	authRepo := postgres.NewAuthRepository(s.db)
	authService := auth.NewService(authRepo, smsSender, jwtManager, rateLimiter)
	authHandler := auth.NewHandler(authService, jwtManager)

	groupRepo := postgres.NewGroupRepository(s.db)
	groupService := groups.NewService(groupRepo)
	groupHandler := groups.NewHandler(groupService)

	feedRepo := postgres.NewFeedRepository(s.db)
	feedService := feed.NewService(feedRepo, groupRepo)
	feedHandler := feed.NewHandler(feedService)

	minioClient, err := minio.New(s.cfg.S3)
	if err != nil {
		s.log.Warn("MinIO unavailable, file uploads disabled", slog.String("err", err.Error()))
	}
	var minioSigner assignments.Signer
	var minioStore assignments.ObjectStore
	if minioClient != nil {
		minioSigner = minioClient
		minioStore = minioClient
	} else {
		minioSigner = noopSigner{}
		minioStore = noopStore{}
	}
	assignRepo := postgres.NewAssignmentRepository(s.db)
	assignService := assignments.NewService(assignRepo, groupRepo, minioSigner, minioStore)
	assignHandler := assignments.NewHandler(assignService)

	quizRepo := postgres.NewQuizRepository(s.db)
	quizService := quizzes.NewService(quizRepo, groupRepo)
	quizHandler := quizzes.NewHandler(quizService)

	gradeRepo := postgres.NewGradeRepository(s.db)
	gradeService := grades.NewService(gradeRepo, groupRepo)
	gradeHandler := grades.NewHandler(gradeService)

	// --- Маршруты ---
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/version", s.handleVersion)

		authHandler.RegisterRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))
			groupHandler.RegisterRoutes(r)
			feedHandler.RegisterRoutes(r)
			assignHandler.RegisterRoutes(r)
			quizHandler.RegisterRoutes(r)
			gradeHandler.RegisterRoutes(r)
		})
	})

	s.router = r
}

// noopSigner and noopStore are used when MinIO is unavailable (e.g., local dev without S3_ENDPOINT).
type noopSigner struct{}

func (noopSigner) PresignedGetURL(_ context.Context, key string) (string, error) { return key, nil }

type noopStore struct{}

func (noopStore) PutObject(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error {
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{"server": "ok", "database": "ok", "redis": "ok"}
	healthy := true

	if err := s.db.PingContext(r.Context()); err != nil {
		checks["database"] = "fail"
		healthy = false
	}
	if err := s.redis.Ping(r.Context()).Err(); err != nil {
		checks["redis"] = "fail"
		healthy = false
	}

	status := "ok"
	if !healthy {
		status = "degraded"
	}
	body := map[string]any{"status": status, "checks": checks}
	if !healthy {
		response.Status(w, http.StatusServiceUnavailable, body)
		return
	}
	response.OK(w, body)
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]any{"version": s.cfg.Server.Version, "env": s.cfg.Server.Env})
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
