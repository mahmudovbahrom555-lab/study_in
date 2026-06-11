package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	mw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

// DocumentUploader is the MinIO upload interface required by the handler.
type DocumentUploader interface {
	PutObject(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
}

// QueueEnqueuer submits background AI tasks to Asynq.
type QueueEnqueuer interface {
	EnqueueAIProcessDocument(docID, jobID string) error
	EnqueueAIGenerateQuiz(p GenerateQuizJobPayload) error
}

// GenerateQuizJobPayload carries data for the async quiz-generation task.
type GenerateQuizJobPayload struct {
	DocumentID   string
	GroupID      string
	TeacherID    string
	JobID        string
	Title        string
	NumQuestions int
	CEFRLevel    string
	Subject      string
}

// Handler exposes AI endpoints under /ai/...
type Handler struct {
	svc     *Service
	upload  DocumentUploader
	enqueue QueueEnqueuer
	maxSize int64
}

func NewHandler(svc *Service, upload DocumentUploader, enqueue QueueEnqueuer, maxDocSize int64) *Handler {
	return &Handler{svc: svc, upload: upload, enqueue: enqueue, maxSize: maxDocSize}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/ai", func(r chi.Router) {
		r.Post("/documents", h.uploadDocument)
		r.Get("/documents", h.listDocuments)
		r.Delete("/documents/{docID}", h.deleteDocument)

		r.Post("/documents/{docID}/generate-quiz", h.generateQuiz)
		r.Get("/jobs/{jobID}", h.getJob)

		r.Post("/sessions", h.createSession)
		r.Get("/sessions/{id}", h.getSession)
		r.Post("/sessions/{id}/messages", h.sendMessage)

		r.Post("/writing/check", h.checkWriting)

		r.Get("/me/gamification", h.getGamification)
		r.Get("/me/weaknesses", h.getWeaknesses)
		r.Get("/me/review-queue", h.getReviewQueue)
	})
}

// ─── Documents ────────────────────────────────────────────────────────────────

func (h *Handler) uploadDocument(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers can upload documents", domain.ErrForbidden))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxSize)
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		response.Error(w, domain.NewError("VALIDATION", "file too large or invalid form", domain.ErrValidation))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "file field required", domain.ErrValidation))
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}

	var groupID *uuid.UUID
	if gid := r.FormValue("group_id"); gid != "" {
		parsed, parseErr := uuid.Parse(gid)
		if parseErr != nil {
			response.Error(w, domain.NewError("VALIDATION", "invalid group_id", domain.ErrValidation))
			return
		}
		groupID = &parsed
	}

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "read file error", domain.ErrInternal))
		return
	}

	objectKey := "ai-docs/" + uuid.New().String() + "/" + header.Filename
	mime := header.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}

	if h.upload != nil {
		if err = h.upload.PutObject(r.Context(), objectKey, bytes.NewReader(data), int64(len(data)), mime); err != nil {
			response.Error(w, domain.NewError("INTERNAL", "upload failed", domain.ErrInternal))
			return
		}
	}

	doc := &domain.AIDocument{
		GroupID:   groupID,
		TeacherID: callerID,
		Title:     title,
		ObjectKey: objectKey,
		MimeType:  mime,
		SizeBytes: int64(len(data)),
	}
	if err = h.svc.CreateDocument(r.Context(), doc); err != nil {
		response.Error(w, domain.NewError("INTERNAL", "create document failed", domain.ErrInternal))
		return
	}

	job, err := h.svc.CreateJob(r.Context(), "process_document", callerID, &doc.ID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "create job failed", domain.ErrInternal))
		return
	}
	if h.enqueue != nil {
		_ = h.enqueue.EnqueueAIProcessDocument(doc.ID.String(), job.ID.String())
	}

	response.Created(w, map[string]any{
		"document": doc,
		"job_id":   job.ID,
	})
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())

	var groupID *uuid.UUID
	if gid := r.URL.Query().Get("group_id"); gid != "" {
		parsed, err := uuid.Parse(gid)
		if err != nil {
			response.Error(w, domain.NewError("VALIDATION", "invalid group_id", domain.ErrValidation))
			return
		}
		groupID = &parsed
	}

	docs, err := h.svc.ListDocuments(r.Context(), callerID, groupID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "list failed", domain.ErrInternal))
		return
	}
	response.OK(w, docs)
}

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	docID, err := uuid.Parse(chi.URLParam(r, "docID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid doc id", domain.ErrValidation))
		return
	}
	if err = h.svc.DeleteDocument(r.Context(), docID, callerID); err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

// ─── Quiz generation ──────────────────────────────────────────────────────────

func (h *Handler) generateQuiz(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers can generate quizzes", domain.ErrForbidden))
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "docID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid doc id", domain.ErrValidation))
		return
	}

	var req struct {
		GroupID      string `json:"group_id"`
		Title        string `json:"title"`
		NumQuestions int    `json:"num_questions"`
		CEFRLevel    string `json:"cefr_level"`
		Subject      string `json:"subject"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid body", domain.ErrValidation))
		return
	}

	groupID, err := uuid.Parse(req.GroupID)
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid group_id", domain.ErrValidation))
		return
	}

	job, err := h.svc.CreateJob(r.Context(), "generate_quiz", callerID, &docID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "create job failed", domain.ErrInternal))
		return
	}

	numQ := req.NumQuestions
	if numQ <= 0 {
		numQ = 20
	}

	if h.enqueue != nil {
		_ = h.enqueue.EnqueueAIGenerateQuiz(GenerateQuizJobPayload{
			DocumentID:   docID.String(),
			GroupID:      groupID.String(),
			TeacherID:    callerID.String(),
			JobID:        job.ID.String(),
			Title:        req.Title,
			NumQuestions: numQ,
			CEFRLevel:    req.CEFRLevel,
			Subject:      req.Subject,
		})
	}

	response.OK(w, map[string]any{
		"job_id": job.ID,
		"status": job.Status,
	})
}

// ─── Job status ───────────────────────────────────────────────────────────────

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "jobID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid job id", domain.ErrValidation))
		return
	}
	job, err := h.svc.GetJob(r.Context(), id)
	if err != nil || job == nil {
		response.Error(w, domain.ErrNotFound)
		return
	}
	response.OK(w, job)
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (h *Handler) createSession(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())

	var req struct {
		Type       string  `json:"type"`
		DocumentID *string `json:"document_id"`
		Subject    *string `json:"subject"`
		CEFRLevel  *string `json:"cefr_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid body", domain.ErrValidation))
		return
	}

	sess := &domain.AISession{
		UserID:      callerID,
		SessionType: req.Type,
		Subject:     req.Subject,
		CEFRLevel:   req.CEFRLevel,
		Metadata:    []byte("{}"),
	}
	if req.DocumentID != nil {
		did, err := uuid.Parse(*req.DocumentID)
		if err == nil {
			sess.DocumentID = &did
		}
	}
	if err := h.svc.CreateSession(r.Context(), sess); err != nil {
		response.Error(w, domain.NewError("INTERNAL", "create session failed", domain.ErrInternal))
		return
	}
	response.Created(w, sess)
}

func (h *Handler) getSession(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid session id", domain.ErrValidation))
		return
	}

	sess, err := h.svc.GetSession(r.Context(), id, callerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	msgs, err := h.svc.GetSessionMessages(r.Context(), id, callerID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, map[string]any{"session": sess, "messages": msgs})
}

func (h *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid session id", domain.ErrValidation))
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		response.Error(w, domain.NewError("VALIDATION", "content required", domain.ErrValidation))
		return
	}

	msg, err := h.svc.SendMentorMessage(r.Context(), SendMessageRequest{
		SessionID: id,
		UserID:    callerID,
		Content:   req.Content,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, msg)
}

// ─── Writing ──────────────────────────────────────────────────────────────────

func (h *Handler) checkWriting(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Text == "" {
		response.Error(w, domain.NewError("VALIDATION", "text required", domain.ErrValidation))
		return
	}

	result, err := h.svc.CheckWriting(r.Context(), WritingCheckRequest{
		UserID: callerID,
		Text:   req.Text,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.OK(w, result)
}

// ─── Gamification ─────────────────────────────────────────────────────────────

func (h *Handler) getGamification(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	g, err := h.svc.GetGamification(r.Context(), callerID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch gamification failed", domain.ErrInternal))
		return
	}
	response.OK(w, g)
}

func (h *Handler) getWeaknesses(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	topics, err := h.svc.GetWeakTopics(r.Context(), callerID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch weaknesses failed", domain.ErrInternal))
		return
	}
	response.OK(w, topics)
}

func (h *Handler) getReviewQueue(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	queue, err := h.svc.GetReviewQueue(r.Context(), callerID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch review queue failed", domain.ErrInternal))
		return
	}
	response.OK(w, queue)
}
