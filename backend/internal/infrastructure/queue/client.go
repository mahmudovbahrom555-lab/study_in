// Package queue wraps Asynq for async task processing.
package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

// Task type constants — all async jobs in the system.
const (
	TypeSMSSend          = "sms:send"
	TypePushNotification = "push:send"
	TypeAIProcessDoc     = "ai:process_document"
	TypeAIGenerateQuiz   = "ai:generate_quiz"
)

// Client enqueues tasks into Redis-backed Asynq queues.
type Client struct {
	c *asynq.Client
}

// NewClient creates an Asynq client backed by the given Redis address.
func NewClient(redisAddr, redisPassword string) *Client {
	return &Client{
		c: asynq.NewClient(asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		}),
	}
}

// Close releases the underlying client connection.
func (c *Client) Close() error { return c.c.Close() }

// EnqueueSMS schedules an SMS delivery task with up to 3 retries.
func (c *Client) EnqueueSMS(ctx context.Context, phone, message string) error {
	payload, err := json.Marshal(SMSPayload{Phone: phone, Message: message})
	if err != nil {
		return fmt.Errorf("queue EnqueueSMS marshal: %w", err)
	}
	_, err = c.c.EnqueueContext(ctx,
		asynq.NewTask(TypeSMSSend, payload),
		asynq.MaxRetry(3),
		asynq.Queue("critical"),
	)
	if err != nil {
		return fmt.Errorf("queue EnqueueSMS: %w", err)
	}
	return nil
}

// EnqueuePush schedules a push notification with up to 2 retries.
func (c *Client) EnqueuePush(ctx context.Context, p PushPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("queue EnqueuePush marshal: %w", err)
	}
	_, err = c.c.EnqueueContext(ctx,
		asynq.NewTask(TypePushNotification, payload),
		asynq.MaxRetry(2),
		asynq.Queue("default"),
	)
	if err != nil {
		return fmt.Errorf("queue EnqueuePush: %w", err)
	}
	return nil
}

// EnqueueAIProcessDocument schedules PDF chunking and embedding for a document.
func (c *Client) EnqueueAIProcessDocument(docID, jobID string) error {
	payload, err := json.Marshal(AIProcessDocPayload{DocumentID: docID, JobID: jobID})
	if err != nil {
		return fmt.Errorf("queue EnqueueAIProcessDocument marshal: %w", err)
	}
	_, err = c.c.Enqueue(
		asynq.NewTask(TypeAIProcessDoc, payload),
		asynq.MaxRetry(2),
		asynq.Queue("low"),
	)
	return err
}

// EnqueueAIGenerateQuiz schedules a GPT quiz generation job.
func (c *Client) EnqueueAIGenerateQuiz(p AIGenerateQuizPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("queue EnqueueAIGenerateQuiz marshal: %w", err)
	}
	_, err = c.c.Enqueue(
		asynq.NewTask(TypeAIGenerateQuiz, payload),
		asynq.MaxRetry(2),
		asynq.Queue("low"),
	)
	return err
}

// SMSPayload is the task payload for TypeSMSSend.
type SMSPayload struct {
	Phone   string `json:"phone"`
	Message string `json:"message"`
}

// PushPayload is the task payload for TypePushNotification.
type PushPayload struct {
	Token string `json:"token"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Data  any    `json:"data,omitempty"`
}

// AIProcessDocPayload carries data for TypeAIProcessDoc tasks.
type AIProcessDocPayload struct {
	DocumentID string `json:"document_id"`
	JobID      string `json:"job_id"`
}

// AIGenerateQuizPayload carries data for TypeAIGenerateQuiz tasks.
type AIGenerateQuizPayload struct {
	DocumentID   string `json:"document_id"`
	GroupID      string `json:"group_id"`
	TeacherID    string `json:"teacher_id"`
	JobID        string `json:"job_id"`
	Title        string `json:"title"`
	NumQuestions int    `json:"num_questions"`
	CEFRLevel    string `json:"cefr_level"`
	Subject      string `json:"subject"`
}
