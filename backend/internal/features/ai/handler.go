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
		// Documents
		r.Post("/documents", h.uploadDocument)
		r.Get("/documents", h.listDocuments)
		r.Delete("/documents/{docID}", h.deleteDocument)

		// Quiz generation (async)
		r.Post("/documents/{docID}/generate-quiz", h.generateQuiz)
		r.Get("/jobs/{jobID}", h.getJob)

		// Quiz question feedback (Teacher Acceptance Rate)
		r.Post("/questions/{questionID}/feedback", h.questionFeedback)
		r.Get("/me/acceptance-rate", h.myAcceptanceRate)
		r.Get("/me/generation-stats", h.generationStats)

		// AI Mentor sessions
		r.Post("/sessions", h.createSession)
		r.Get("/sessions/{id}", h.getSession)
		r.Post("/sessions/{id}/messages", h.sendMessage)

		// Writing checker
		r.Post("/writing/check", h.checkWriting)

		// Gamification + spaced repetition
		r.Get("/me/gamification", h.getGamification)
		r.Get("/me/weaknesses", h.getWeaknesses)
		r.Get("/me/review-queue", h.getReviewQueue)

		// Recommendations (Phase 3)
		r.Post("/recommendations/{recID}/action", h.recommendationAction)
		r.Get("/recommendations/{recID}/explain", h.recommendationExplain)

		// Demo group seeder (Phase 4)
		r.Post("/me/demo", h.createDemoGroup)
	})

	// Teacher Dashboard — mounted under /groups/:groupID/ai-insights
	r.Get("/groups/{groupID}/ai-insights", h.classInsights)
	r.Get("/groups/{groupID}/students/{studentID}/progress", h.studentProgress)
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

// ─── Quiz question feedback ───────────────────────────────────────────────────

func (h *Handler) questionFeedback(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers can rate questions", domain.ErrForbidden))
		return
	}
	qID, err := uuid.Parse(chi.URLParam(r, "questionID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid question id", domain.ErrValidation))
		return
	}
	var req struct {
		Accepted bool `json:"accepted"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid body", domain.ErrValidation))
		return
	}
	if err = h.svc.SubmitQuestionFeedback(r.Context(), qID, callerID, req.Accepted); err != nil {
		response.Error(w, domain.NewError("INTERNAL", "save feedback failed", domain.ErrInternal))
		return
	}
	response.NoContent(w)
}

func (h *Handler) myAcceptanceRate(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	rate, total, err := h.svc.GetMyAcceptanceRate(r.Context(), callerID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch acceptance rate failed", domain.ErrInternal))
		return
	}
	response.OK(w, map[string]any{"acceptance_rate": rate, "total_questions": total})
}

func (h *Handler) generationStats(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	stats, err := h.svc.GetTeacherStats(r.Context(), callerID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch generation stats failed", domain.ErrInternal))
		return
	}
	response.OK(w, stats)
}

// ─── Teacher Dashboard ────────────────────────────────────────────────────────

func (h *Handler) classInsights(w http.ResponseWriter, r *http.Request) {
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers can view insights", domain.ErrForbidden))
		return
	}
	groupID, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid group id", domain.ErrValidation))
		return
	}
	teacherID := mw.UserIDFromCtx(r.Context())
	insights, err := h.svc.GetClassInsights(r.Context(), groupID, teacherID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch insights failed", domain.ErrInternal))
		return
	}
	response.OK(w, insights)
}

func (h *Handler) studentProgress(w http.ResponseWriter, r *http.Request) {
	callerID := mw.UserIDFromCtx(r.Context())
	callerRole := mw.RoleFromCtx(r.Context())
	if callerRole != domain.RoleTeacher && callerRole != "parent" {
		response.Error(w, domain.NewError("FORBIDDEN", "access denied", domain.ErrForbidden))
		return
	}
	groupID, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid group id", domain.ErrValidation))
		return
	}
	studentID, err := uuid.Parse(chi.URLParam(r, "studentID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid student id", domain.ErrValidation))
		return
	}
	_ = callerID // access control delegated to service layer in full impl

	// Demo student IDs are synthetic — no real DB records exist for them.
	if IsDemoStudentID(studentID) {
		response.OK(w, demoProgress(studentID))
		return
	}

	progress, err := h.svc.GetStudentProgress(r.Context(), groupID, studentID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "fetch progress failed", domain.ErrInternal))
		return
	}
	response.OK(w, progress)
}

func (h *Handler) recommendationAction(w http.ResponseWriter, r *http.Request) {
	teacherID := mw.UserIDFromCtx(r.Context())
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers", domain.ErrForbidden))
		return
	}
	recID, err := uuid.Parse(chi.URLParam(r, "recID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid recommendation id", domain.ErrValidation))
		return
	}
	var body struct {
		Status string `json:"status"` // accepted | dismissed | snoozed
		Action string `json:"action"` // free-text: what the teacher plans/did
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid body", domain.ErrValidation))
		return
	}
	if body.Status != "accepted" && body.Status != "dismissed" && body.Status != "snoozed" {
		response.Error(w, domain.NewError("VALIDATION", "status must be accepted|dismissed|snoozed", domain.ErrValidation))
		return
	}
	// Demo recommendation IDs are synthetic — no DB record to update.
	if IsDemoRecID(recID) {
		response.OK(w, map[string]string{"status": body.Status})
		return
	}
	if err = h.svc.RecordRecommendationAction(r.Context(), recID, teacherID, body.Status, body.Action); err != nil {
		response.Error(w, domain.NewError("INTERNAL", "record action failed", domain.ErrInternal))
		return
	}
	response.OK(w, map[string]string{"status": body.Status})
}

func (h *Handler) recommendationExplain(w http.ResponseWriter, r *http.Request) {
	teacherID := mw.UserIDFromCtx(r.Context())
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers", domain.ErrForbidden))
		return
	}
	recID, err := uuid.Parse(chi.URLParam(r, "recID"))
	if err != nil {
		response.Error(w, domain.NewError("VALIDATION", "invalid recommendation id", domain.ErrValidation))
		return
	}
	// Demo recommendations have no DB record — return the hardcoded reason directly.
	if IsDemoRecID(recID) {
		response.OK(w, map[string]string{"explanation": demoRecExplanation(recID)})
		return
	}
	text, err := h.svc.ExplainRecommendation(r.Context(), recID, teacherID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "explain failed", domain.ErrInternal))
		return
	}
	response.OK(w, map[string]string{"explanation": text})
}

func (h *Handler) createDemoGroup(w http.ResponseWriter, r *http.Request) {
	teacherID := mw.UserIDFromCtx(r.Context())
	if mw.RoleFromCtx(r.Context()) != domain.RoleTeacher {
		response.Error(w, domain.NewError("FORBIDDEN", "only teachers", domain.ErrForbidden))
		return
	}
	groupID, err := h.svc.SeedDemoGroup(r.Context(), teacherID)
	if err != nil {
		response.Error(w, domain.NewError("INTERNAL", "demo group creation failed", domain.ErrInternal))
		return
	}
	response.OK(w, map[string]string{"group_id": groupID.String()})
}

// ─── Demo helpers ─────────────────────────────────────────────────────────────

// demoProgress returns a stub StudentProgress for a demo student so that the
// detail page doesn't 500 when a teacher clicks through from the demo group.
func demoProgress(studentID uuid.UUID) *domain.StudentProgress {
	names := map[uuid.UUID]string{
		demoStudentIDs[0]: "Алибек Жумаев",
		demoStudentIDs[1]: "Малика Рашидова",
		demoStudentIDs[2]: "Дилноза Каримова",
		demoStudentIDs[3]: "Жасурбек Ибрагимов",
		demoStudentIDs[4]: "Нилуфар Юсупова",
		demoStudentIDs[5]: "Отабек Хасанов",
	}
	return &domain.StudentProgress{
		StudentID: studentID,
		Name:      names[studentID],
	}
}

// demoRecExplanation returns a static explanation for each demo recommendation.
func demoRecExplanation(recID uuid.UUID) string {
	explanations := map[uuid.UUID]string{
		demoRecIDs[0]: "Passive Voice — самая распространённая грамматическая ошибка на уровне B2. " +
			"4 из 6 учеников показывают точность ниже 40% при стабильно низкой уверенности (31%). " +
			"Рекомендуем создать короткий тест из 10 вопросов с фокусом на Present/Past Passive.",
		demoRecIDs[1]: "2 ученика не открывали приложение более 4 дней подряд. " +
			"По данным платформы, пропуск более 3 дней увеличивает вероятность отчисления на 40%. " +
			"Простое сообщение с поддержкой возвращает 60% неактивных студентов.",
		demoRecIDs[2]: "Conditionals показывают нестабильный прогресс: 3 ученика знают правило, " +
			"но делают ошибки в использовании (consistency 40%). " +
			"Разбор в классе с практическими примерами из реальной речи даст быстрый результат.",
		demoRecIDs[3]: "TAR (Teacher Acceptance Rate) 71% означает, что каждый 3-й вопрос AI " +
			"не соответствует вашим стандартам. Оценка вопросов помогает AI обучиться " +
			"вашему стилю и улучшить качество до 85%+ за 2-3 недели.",
	}
	if exp, ok := explanations[recID]; ok {
		return exp
	}
	return "Это демонстрационная рекомендация. В реальной группе здесь появится персональный анализ от AI."
}
