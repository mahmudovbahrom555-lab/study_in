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
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	redisclient "github.com/redis/go-redis/v9"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/config"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	aifeature "github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/ai"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/assignments"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/attendance"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/auth"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/feed"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/grades"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/groups"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/notifications"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/parents"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/quizzes"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/reports"
	openaiinfra "github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/openai"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/minio"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/postgres"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/queue"
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
	aiSvc  *aifeature.Service
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

	// Async task queue — SMS and push go through Asynq (backed by Redis).
	queueClient := queue.NewClient(s.cfg.Redis.Addr, s.cfg.Redis.Password)

	var directSMS sms.Sender
	if s.cfg.SMS.Provider == "eskiz" {
		directSMS = sms.NewEskizSender(s.cfg.SMS.EskizEmail, s.cfg.SMS.EskizPassword, s.cfg.SMS.EskizFrom)
	} else {
		directSMS = sms.NewMockSender()
	}
	// Worker picks up SMS tasks from the queue and calls directSMS.Send.
	// In server we only enqueue — actual delivery is async.
	smsSender := &asyncSMSSender{q: queueClient, fallback: directSMS}

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
	// observer wired after aiService is constructed below
	quizService := quizzes.NewService(quizRepo, groupRepo, nil)
	quizHandler := quizzes.NewHandler(quizService)

	gradeRepo := postgres.NewGradeRepository(s.db)
	gradeService := grades.NewService(gradeRepo, groupRepo)
	gradeHandler := grades.NewHandler(gradeService)

	attendanceRepo := postgres.NewAttendanceRepository(s.db)
	attendanceService := attendance.NewService(attendanceRepo, groupRepo)
	attendanceHandler := attendance.NewHandler(attendanceService)

	parentRepo := postgres.NewParentRepository(s.db)
	parentService := parents.NewService(parentRepo, gradeRepo, attendanceRepo, groupMemberAdapter{groupRepo})
	parentHandler := parents.NewHandler(parentService)

	reportRepo := postgres.NewReportRepository(s.db)
	reportService := reports.NewService(reportRepo)
	reportHandler := reports.NewHandler(reportService, reportRepo)

	notifRepo := postgres.NewNotificationRepository(s.db)
	notifService := notifications.NewService(notifRepo, &pushEnqueuer{q: queueClient})
	notifHandler := notifications.NewHandler(notifService)

	// ── AI layer ────────────────────────────────────────────────────────────────
	openaiClient := openaiinfra.NewClient(
		s.cfg.AI.OpenAI.APIKey,
		s.cfg.AI.OpenAI.Model,
		s.cfg.AI.OpenAI.EmbeddingModel,
	)
	aiDocRepo := postgres.NewAIDocumentRepository(s.db)
	aiJobRepo := postgres.NewAIJobRepository(s.db)
	aiSessRepo := postgres.NewAISessionRepository(s.db)
	aiTokenRepo := postgres.NewAITokenRepository(s.db)
	aiMasteryRepo := postgres.NewAIMasteryRepository(s.db)
	aiGamifRepo := postgres.NewAIGamificationRepository(s.db)
	aiTopicRepo := postgres.NewAITopicRepository(s.db)
	aiQuizCreator := &aiQuizCreatorAdapter{quizRepo: quizRepo, topicRepo: aiTopicRepo}
	aiInsightsRepo := postgres.NewInsightsRepository(s.db)
	aiFeedbackRepo := postgres.NewQuizFeedbackRepository(s.db)
	aiGenSessRepo := postgres.NewAIGenerationSessionRepository(s.db)
	aiRecRepo := postgres.NewRecommendationRepository(s.db)

	var aiObjectStore aifeature.ObjectStore
	if minioClient != nil {
		aiObjectStore = minioClient
	} else {
		aiObjectStore = noopAIStore{}
	}

	aiDemoRepo := postgres.NewDemoRepository(s.db)
	s.aiSvc = aifeature.NewService(
		aiDocRepo, aiJobRepo, aiSessRepo, aiTokenRepo,
		aiMasteryRepo, aiGamifRepo, aiTopicRepo, aiQuizCreator, aiObjectStore,
		aiInsightsRepo, aiFeedbackRepo, aiGenSessRepo, aiRecRepo, aiDemoRepo,
		openaiClient,
		aifeature.ServiceConfig{
			Model:            s.cfg.AI.OpenAI.Model,
			MonthlyTokensMax: s.cfg.AI.OpenAI.MonthlyTokensMax,
			ChunkSize:        s.cfg.AI.ChunkSize,
			ChunkOverlap:     s.cfg.AI.ChunkOverlap,
		},
	)

	// Wire SM-2 observer now that aiService is ready.
	quizService.SetObserver(s.aiSvc)

	var aiUploader aifeature.DocumentUploader
	if minioClient != nil {
		aiUploader = minioClient
	}
	aiHandler := aifeature.NewHandler(s.aiSvc, aiUploader, &aiQueueEnqueuerAdapter{q: queueClient}, s.cfg.AI.MaxDocSizeBytes)

	// --- Маршруты ---
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", s.handleHealth)
		r.Get("/version", s.handleVersion)

		authHandler.RegisterRoutes(r)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtManager))
			r.Use(middleware.Metrics)
			groupHandler.RegisterRoutes(r)
			feedHandler.RegisterRoutes(r)
			assignHandler.RegisterRoutes(r)
			quizHandler.RegisterRoutes(r)
			gradeHandler.RegisterRoutes(r)
			attendanceHandler.RegisterRoutes(r)
			parentHandler.RegisterRoutes(r)
			notifHandler.RegisterRoutes(r)
			aiHandler.RegisterRoutes(r)
			reportHandler.RegisterRoutes(r)
		})
	})

	s.router = r
}

// groupMemberAdapter adapts groups.Repository to parents.GroupMemberLister.
type groupMemberAdapter struct{ r groups.Repository }

func (a groupMemberAdapter) ListGroupsByStudent(ctx context.Context, studentID uuid.UUID) ([]*domain.Group, error) {
	return a.r.ListStudentGroups(ctx, studentID)
}

// noopSigner and noopStore are used when MinIO is unavailable (e.g., local dev without S3_ENDPOINT).
type noopSigner struct{}

func (noopSigner) PresignedGetURL(_ context.Context, key string) (string, error) { return key, nil }

type noopStore struct{}

func (noopStore) PutObject(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error {
	return nil
}

// asyncSMSSender satisfies sms.Sender by enqueuing tasks via Asynq.
// Falls back to directSMS when the queue is unavailable (dev/test).
type asyncSMSSender struct {
	q        interface {
		EnqueueSMS(ctx context.Context, phone, message string) error
	}
	fallback sms.Sender
}

func (a *asyncSMSSender) Send(ctx context.Context, phone, message string) error {
	if err := a.q.EnqueueSMS(ctx, phone, message); err != nil {
		// Queue unavailable — fall back to synchronous send.
		return a.fallback.Send(ctx, phone, message)
	}
	return nil
}

// pushEnqueuer satisfies notifications.TaskEnqueuer via queue.Client.
type pushEnqueuer struct{ q *queue.Client }

func (p *pushEnqueuer) EnqueuePush(ctx context.Context, token, title, body string, data any) error {
	return p.q.EnqueuePush(ctx, queue.PushPayload{Token: token, Title: title, Body: body, Data: data})
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

// AIProcessor returns a queue.AIProcessor adapter backed by the AI service.
// Returns nil when the AI service was not initialised (no DB / no setup).
func (s *Server) AIProcessor() queue.AIProcessor {
	if s.aiSvc == nil {
		return nil
	}
	return &aiProcessorAdapter{svc: s.aiSvc}
}

// ─── AI adapters ──────────────────────────────────────────────────────────────

// aiProcessorAdapter satisfies queue.AIProcessor using the AI service.
type aiProcessorAdapter struct{ svc *aifeature.Service }

func (a *aiProcessorAdapter) ProcessDocument(ctx context.Context, docID uuid.UUID) error {
	return a.svc.ProcessDocument(ctx, docID)
}

func (a *aiProcessorAdapter) GenerateQuizByPayload(ctx context.Context, p queue.AIGenerateQuizPayload) error {
	return a.svc.GenerateQuizFromPayload(ctx,
		p.DocumentID, p.GroupID, p.TeacherID, p.JobID,
		p.Title, p.NumQuestions, p.CEFRLevel, p.Subject,
	)
}

// aiQueueEnqueuerAdapter satisfies aifeature.QueueEnqueuer using queue.Client.
type aiQueueEnqueuerAdapter struct{ q *queue.Client }

func (e *aiQueueEnqueuerAdapter) EnqueueAIProcessDocument(docID, jobID string) error {
	return e.q.EnqueueAIProcessDocument(docID, jobID)
}

func (e *aiQueueEnqueuerAdapter) EnqueueAIGenerateQuiz(p aifeature.GenerateQuizJobPayload) error {
	return e.q.EnqueueAIGenerateQuiz(queue.AIGenerateQuizPayload{
		DocumentID:   p.DocumentID,
		GroupID:      p.GroupID,
		TeacherID:    p.TeacherID,
		JobID:        p.JobID,
		Title:        p.Title,
		NumQuestions: p.NumQuestions,
		CEFRLevel:    p.CEFRLevel,
		Subject:      p.Subject,
	})
}

// aiQuizCreatorAdapter satisfies aifeature.QuizCreator using quizzes.Repository.
type aiQuizCreatorAdapter struct {
	quizRepo  quizzes.Repository
	topicRepo *postgres.AITopicRepository
}

func (c *aiQuizCreatorAdapter) CreateQuizWithQuestions(ctx context.Context, req aifeature.QuizCreateRequest) (uuid.UUID, error) {
	quizID := uuid.New()
	now := time.Now()
	q := &domain.Quiz{
		ID:          quizID,
		GroupID:     req.GroupID,
		TeacherID:   req.TeacherID,
		Title:       req.Title,
		MaxAttempts: 1,
		IsPublished: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := c.quizRepo.CreateQuiz(ctx, q); err != nil {
		return uuid.Nil, err
	}
	for i, gq := range req.Questions {
		qID := uuid.New()
		exp := gq.Explanation
		question := &domain.Question{
			ID:          qID,
			QuizID:      quizID,
			Body:        gq.Text,
			Explanation: &exp,
			Position:    int16(i + 1),
			Points:      1,
		}
		if err := c.quizRepo.CreateQuestion(ctx, question); err != nil {
			return uuid.Nil, err
		}
		for j, opt := range gq.Options {
			o := &domain.Option{
				ID:         uuid.New(),
				QuestionID: qID,
				Body:       opt.Text,
				IsCorrect:  opt.IsCorrect,
				Position:   int16(j + 1),
			}
			if err := c.quizRepo.CreateOption(ctx, o); err != nil {
				return uuid.Nil, err
			}
		}
		for _, tagName := range gq.TopicTags {
			tag, err := c.topicRepo.UpsertTopic(ctx, tagName, nil)
			if err == nil && tag != nil {
				_ = c.topicRepo.LinkQuestionTopics(ctx, qID, []uuid.UUID{tag.ID})
			}
		}
	}
	return quizID, nil
}

// noopAIStore satisfies aifeature.ObjectStore when MinIO is unavailable.
type noopAIStore struct{}

func (noopAIStore) GetObject(_ context.Context, _ string) ([]byte, error) { return nil, nil }

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
