package assignments

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

const maxUploadMemory = 32 << 20 // 32 MB multipart buffer

type Handler struct {
	svc      *Service
	validate *validator.Validate
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc, validate: validator.New()}
}

// RegisterRoutes mounts all assignment routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/groups/{groupID}/assignments", func(r chi.Router) {
		r.Get("/", h.listAssignments)
		r.Post("/", h.createAssignment)

		r.Route("/{assignmentID}", func(r chi.Router) {
			r.Get("/", h.getAssignment)
			r.Patch("/", h.updateAssignment)
			r.Delete("/", h.deleteAssignment)
			r.Post("/files", h.uploadAssignmentFile)

			r.Post("/submit", h.submit)
			r.Post("/submit/files", h.uploadSubmissionFile)

			r.Get("/submissions", h.listSubmissions)
			r.Post("/submissions/{submissionID}/grade", h.gradeSubmission)
		})
	})
}

// GET /groups/{groupID}/assignments
func (h *Handler) listAssignments(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	as, err := h.svc.ListAssignments(r.Context(), groupID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]AssignmentResponse, len(as))
	for i, a := range as {
		rawAtts, _ := h.svc.AssignmentAttachments(r.Context(), a.ID)
		out[i] = assignmentToResponse(a, h.buildAttachments(r, rawAtts))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /groups/{groupID}/assignments
func (h *Handler) createAssignment(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, chi.URLParam(r, "groupID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	if role != domain.RoleTeacher {
		writeError(w, http.StatusForbidden, "FORBIDDEN", "only teachers can create assignments")
		return
	}

	var req CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	a, err := h.svc.CreateAssignment(r.Context(), callerID, groupID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, assignmentToResponse(a, nil))
}

// GET /groups/{groupID}/assignments/{assignmentID}
func (h *Handler) getAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())
	role := apimw.RoleFromCtx(r.Context())

	a, atts, err := h.svc.GetAssignment(r.Context(), assignmentID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assignmentToResponse(a, h.buildAttachments(r, atts)))
}

// PATCH /groups/{groupID}/assignments/{assignmentID}
func (h *Handler) updateAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req UpdateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	a, err := h.svc.UpdateAssignment(r.Context(), assignmentID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, assignmentToResponse(a, nil))
}

// DELETE /groups/{groupID}/assignments/{assignmentID}
func (h *Handler) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := h.svc.DeleteAssignment(r.Context(), assignmentID, callerID); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /groups/{groupID}/assignments/{assignmentID}/files  (multipart)
func (h *Handler) uploadAssignmentFile(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "multipart parse error")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing file field")
		return
	}
	defer file.Close()

	att, err := h.svc.UploadAssignmentFile(r.Context(), assignmentID, callerID, header.Filename, file, header.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	url, _ := h.svc.SignedURL(r.Context(), att.ObjectKey)
	writeJSON(w, http.StatusCreated, AttachmentResponse{
		ID:        att.ID,
		Filename:  att.Filename,
		MimeType:  att.MimeType,
		SizeBytes: att.SizeBytes,
		URL:       url,
	})
}

// POST /groups/{groupID}/assignments/{assignmentID}/submit
func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}

	sub, err := h.svc.Submit(r.Context(), assignmentID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, submissionToResponse(sub, nil))
}

// POST /groups/{groupID}/assignments/{assignmentID}/submit/files
func (h *Handler) uploadSubmissionFile(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	// Determine submissionID for the current student.
	sub, err := h.svc.GetStudentSubmission(r.Context(), assignmentID, callerID)
	if err != nil || sub == nil {
		writeError(w, http.StatusBadRequest, "NO_SUBMISSION", "submit first")
		return
	}

	if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "multipart parse error")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "missing file field")
		return
	}
	defer file.Close()

	att, err := h.svc.UploadSubmissionFile(r.Context(), sub.ID, callerID, header.Filename, file, header.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	url, _ := h.svc.SignedURL(r.Context(), att.ObjectKey)
	writeJSON(w, http.StatusCreated, AttachmentResponse{
		ID:        att.ID,
		Filename:  att.Filename,
		MimeType:  att.MimeType,
		SizeBytes: att.SizeBytes,
		URL:       url,
	})
}

// GET /groups/{groupID}/assignments/{assignmentID}/submissions
func (h *Handler) listSubmissions(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	subs, err := h.svc.ListSubmissions(r.Context(), assignmentID, callerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]SubmissionResponse, len(subs))
	for i, sub := range subs {
		rawAtts, _ := h.svc.SubmissionAttachments(r.Context(), sub.ID)
		out[i] = submissionToResponse(sub, h.buildSubAttachments(r, rawAtts))
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /groups/{groupID}/assignments/{assignmentID}/submissions/{submissionID}/grade
func (h *Handler) gradeSubmission(w http.ResponseWriter, r *http.Request) {
	submissionID, ok := parseUUID(w, chi.URLParam(r, "submissionID"))
	if !ok {
		return
	}
	callerID := apimw.UserIDFromCtx(r.Context())

	var req GradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	sub, err := h.svc.Grade(r.Context(), submissionID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, submissionToResponse(sub, nil))
}

// --- helpers ---

func (h *Handler) buildAttachments(r *http.Request, atts []*domain.AssignmentAttachment) []AttachmentResponse {
	if len(atts) == 0 {
		return nil
	}
	out := make([]AttachmentResponse, len(atts))
	for i, a := range atts {
		url, _ := h.svc.SignedURL(r.Context(), a.ObjectKey)
		out[i] = AttachmentResponse{ID: a.ID, Filename: a.Filename, MimeType: a.MimeType, SizeBytes: a.SizeBytes, URL: url}
	}
	return out
}

func (h *Handler) buildSubAttachments(r *http.Request, atts []*domain.SubmissionAttachment) []AttachmentResponse {
	if len(atts) == 0 {
		return nil
	}
	out := make([]AttachmentResponse, len(atts))
	for i, a := range atts {
		url, _ := h.svc.SignedURL(r.Context(), a.ObjectKey)
		out[i] = AttachmentResponse{ID: a.ID, Filename: a.Filename, MimeType: a.MimeType, SizeBytes: a.SizeBytes, URL: url}
	}
	return out
}

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
	case errors.Is(err, domain.ErrValidation):
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "file too large or invalid")
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
