package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/sms"
)

// Worker runs Asynq task handlers in background goroutines.
type Worker struct {
	srv    *asynq.Server
	mux    *asynq.ServeMux
	log    *slog.Logger
	sender sms.Sender
	pusher Pusher
	ai     AIProcessor
}

// Pusher sends a push notification to a device token.
type Pusher interface {
	Send(ctx context.Context, token, title, body string, data any) error
}

// AIProcessor handles heavy AI tasks (PDF processing, quiz generation).
// Implemented by features/ai.Service; nil disables AI workers.
type AIProcessor interface {
	ProcessDocument(ctx context.Context, docID uuid.UUID) error
	GenerateQuizByPayload(ctx context.Context, p AIGenerateQuizPayload) error
}

// NewWorker creates a worker connected to Redis with the given concurrency.
// ai may be nil — AI task handlers register only when it is non-nil.
func NewWorker(redisAddr, redisPassword string, concurrency int, log *slog.Logger, sender sms.Sender, pusher Pusher, ai AIProcessor) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Error("task failed",
					slog.String("type", task.Type()),
					slog.String("err", err.Error()),
				)
			}),
		},
	)
	w := &Worker{srv: srv, mux: asynq.NewServeMux(), log: log, sender: sender, pusher: pusher, ai: ai}
	w.mux.HandleFunc(TypeSMSSend, w.handleSMS)
	w.mux.HandleFunc(TypePushNotification, w.handlePush)
	if ai != nil {
		w.mux.HandleFunc(TypeAIProcessDoc, w.handleAIProcessDoc)
		w.mux.HandleFunc(TypeAIGenerateQuiz, w.handleAIGenerateQuiz)
	}
	return w
}

// Start runs the worker. Blocks until Stop is called.
func (w *Worker) Start() error {
	w.log.Info("asynq worker starting")
	if err := w.srv.Start(w.mux); err != nil {
		return fmt.Errorf("asynq worker start: %w", err)
	}
	return nil
}

// Stop gracefully shuts down the worker.
func (w *Worker) Stop() {
	w.srv.Shutdown()
	w.log.Info("asynq worker stopped")
}

func (w *Worker) handleSMS(ctx context.Context, task *asynq.Task) error {
	var p SMSPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return fmt.Errorf("handleSMS unmarshal: %w", err)
	}
	if err := w.sender.Send(ctx, p.Phone, p.Message); err != nil {
		return fmt.Errorf("handleSMS send to %s: %w", p.Phone, err)
	}
	w.log.Info("sms sent", slog.String("phone", p.Phone))
	return nil
}

func (w *Worker) handlePush(ctx context.Context, task *asynq.Task) error {
	if w.pusher == nil {
		return nil
	}
	var p PushPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return fmt.Errorf("handlePush unmarshal: %w", err)
	}
	if err := w.pusher.Send(ctx, p.Token, p.Title, p.Body, p.Data); err != nil {
		return fmt.Errorf("handlePush send: %w", err)
	}
	w.log.Info("push sent", slog.String("token_prefix", p.Token[:min(8, len(p.Token))]))
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (w *Worker) handleAIProcessDoc(ctx context.Context, task *asynq.Task) error {
	var p AIProcessDocPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return fmt.Errorf("handleAIProcessDoc unmarshal: %w", err)
	}
	docID, err := uuid.Parse(p.DocumentID)
	if err != nil {
		return fmt.Errorf("handleAIProcessDoc parse docID: %w", err)
	}
	if err = w.ai.ProcessDocument(ctx, docID); err != nil {
		return fmt.Errorf("handleAIProcessDoc process: %w", err)
	}
	w.log.Info("ai document processed", slog.String("doc_id", p.DocumentID))
	return nil
}

func (w *Worker) handleAIGenerateQuiz(ctx context.Context, task *asynq.Task) error {
	var p AIGenerateQuizPayload
	if err := json.Unmarshal(task.Payload(), &p); err != nil {
		return fmt.Errorf("handleAIGenerateQuiz unmarshal: %w", err)
	}
	if err := w.ai.GenerateQuizByPayload(ctx, p); err != nil {
		return fmt.Errorf("handleAIGenerateQuiz generate: %w", err)
	}
	w.log.Info("ai quiz generated", slog.String("doc_id", p.DocumentID), slog.String("job_id", p.JobID))
	return nil
}
