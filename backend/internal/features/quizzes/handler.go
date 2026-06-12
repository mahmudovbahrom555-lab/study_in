package quizzes

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	apimw "github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
)

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups/{groupID}/quizzes", func(r chi.Router) {
		r.Get("/", h.listQuizzes)
		r.Post("/", h.createQuiz)

		r.Route("/{quizID}", func(r chi.Router) {
			r.Get("/", h.getQuiz)
			r.Post("/publish", h.publishQuiz)
			r.Post("/unpublish", h.unpublishQuiz)

			r.Post("/questions", h.addQuestion)
			r.Route("/questions/{questionID}", func(r chi.Router) {
				r.Post("/options", h.addOption)
			})

			r.Post("/attempt", h.startAttempt)
			r.Post("/attempt/{attemptID}/submit", h.submitAttempt)
		})
	})
}

func (h *Handler) listQuizzes(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	qs, err := h.svc.ListGroupQuizzes(r.Context(), groupID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]QuizResponse, len(qs))
	for i, q := range qs {
		out[i] = quizToResponse(q)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) createQuiz(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can create quizzes")
		return
	}

	var req CreateQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	q, err := h.svc.CreateQuiz(r.Context(), callerID, groupID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, quizToResponse(q))
}

func (h *Handler) getQuiz(w http.ResponseWriter, r *http.Request) {
	quizID, ok := parseUUID(w, chi.URLParam(r, "quizID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	q, questions, optMap, err := h.svc.GetQuiz(r.Context(), quizID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	resp := quizToResponse(q)
	isTeacher := role == domain.RoleTeacher

	for _, qst := range questions {
		qstResp := QuestionResponse{
			ID:          qst.ID,
			Body:        qst.Body,
			Explanation: qst.Explanation,
			Position:    qst.Position,
			Points:      qst.Points,
		}
		for _, o := range optMap[qst.ID] {
			or := OptionResponse{ID: o.ID, Body: o.Body, Position: o.Position}
			if isTeacher {
				or.IsCorrect = &o.IsCorrect
			}
			qstResp.Options = append(qstResp.Options, or)
		}
		resp.Questions = append(resp.Questions, qstResp)
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) publishQuiz(w http.ResponseWriter, r *http.Request) {
	h.setPublished(w, r, true)
}

func (h *Handler) unpublishQuiz(w http.ResponseWriter, r *http.Request) {
	h.setPublished(w, r, false)
}

func (h *Handler) setPublished(w http.ResponseWriter, r *http.Request, published bool) {
	quizID, ok := parseUUID(w, chi.URLParam(r, "quizID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	q, err := h.svc.Publish(r.Context(), quizID, callerID, published)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, quizToResponse(q))
}

func (h *Handler) addQuestion(w http.ResponseWriter, r *http.Request) {
	quizID, ok := parseUUID(w, chi.URLParam(r, "quizID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req CreateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	qst, err := h.svc.AddQuestion(r.Context(), quizID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, QuestionResponse{
		ID:       qst.ID,
		Body:     qst.Body,
		Position: qst.Position,
		Points:   qst.Points,
	})
}

func (h *Handler) addOption(w http.ResponseWriter, r *http.Request) {
	questionID, ok := parseUUID(w, chi.URLParam(r, "questionID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req CreateOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	o, err := h.svc.AddOption(r.Context(), questionID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	isCorrect := o.IsCorrect
	writeJSON(w, http.StatusCreated, OptionResponse{
		ID:        o.ID,
		Body:      o.Body,
		Position:  o.Position,
		IsCorrect: &isCorrect,
	})
}

func (h *Handler) startAttempt(w http.ResponseWriter, r *http.Request) {
	quizID, ok := parseUUID(w, chi.URLParam(r, "quizID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	a, err := h.svc.StartAttempt(r.Context(), quizID, callerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, attemptToResponse(a))
}

func (h *Handler) submitAttempt(w http.ResponseWriter, r *http.Request) {
	attemptID, ok := parseUUID(w, chi.URLParam(r, "attemptID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req SubmitAnswersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}

	result, err := h.svc.SubmitAttempt(r.Context(), attemptID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, SubmitAttemptFullResponse{
		Attempt:         attemptToResponse(result.Attempt),
		QuestionResults: result.QuestionResults,
	})
}

// --- helpers ---

func parseUUID(w http.ResponseWriter, s string) (uuid.UUID, bool) {
	id, err := uuid.Parse(s)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid UUID")
		return uuid.Nil, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": code, "message": message},
	})
}

func writeValidationError(w http.ResponseWriter, err error) {
	var ve validator.ValidationErrors
	details := map[string]any{}
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details[fe.Field()] = fe.Tag()
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": "VALIDATION_ERROR", "message": "validation failed", "details": details},
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	case errors.Is(err, domain.ErrForbidden):
		writeError(w, http.StatusForbidden, "FORBIDDEN", "forbidden")
	case errors.Is(err, domain.ErrConflict):
		writeError(w, http.StatusConflict, "CONFLICT", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
