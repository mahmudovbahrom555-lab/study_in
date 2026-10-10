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
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
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
			r.Get("/submission", h.mySubmission)

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

	items, err := h.svc.ListAssignments(r.Context(), groupID, callerID, role)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]AssignmentResponse, len(items))
	for i, it := range items {
		rawAtts, _ := h.svc.AssignmentAttachments(r.Context(), it.Assignment.ID)
		out[i] = assignmentToResponse(it.Assignment, h.buildAttachments(r, rawAtts))
		if s := it.MySubmission; s != nil {
			out[i].MySubmission = &SubmissionSummary{Grade: s.Grade, SubmittedAt: s.SubmittedAt}
		}
		if st := it.Stats; st != nil {
			out[i].Stats = &StatsResponse{Submitted: st.Submitted, Ungraded: st.Ungraded}
		}
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

	url, _ := h.svc.SignedURL(r.Context(), att.ObjectKey, att.Filename)
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
	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, err)
		return
	}

	sub, err := h.svc.Submit(r.Context(), assignmentID, callerID, req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	files, err := h.svc.SubmissionFiles(r.Context(), []uuid.UUID{sub.ID})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, submissionToResponse(sub, h.buildFiles(r, files[sub.ID])))
}

// GET /groups/{groupID}/assignments/{assignmentID}/submission — своя сдача ученика
// (для экрана пересдачи). 404, если ещё не сдавал.
func (h *Handler) mySubmission(w http.ResponseWriter, r *http.Request) {
	assignmentID, ok := parseUUID(w, chi.URLParam(r, "assignmentID"))
	if !ok {
		return
	}
	sub, files, err := h.svc.MySubmission(r.Context(), assignmentID, apimw.UserIDFromCtx(r.Context()))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if sub == nil {
		writeServiceError(w, domain.ErrNotFound)
		return
	}
	writeJSON(w, http.StatusOK, submissionToResponse(sub, h.buildFiles(r, files)))
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

	ids := make([]uuid.UUID, len(subs))
	for i, sub := range subs {
		ids[i] = sub.ID
	}
	// Один запрос на файлы всех сдач, а не по запросу на каждую.
	files, err := h.svc.SubmissionFiles(r.Context(), ids)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	out := make([]SubmissionResponse, len(subs))
	for i, sub := range subs {
		out[i] = submissionToResponse(&sub.Submission, h.buildFiles(r, files[sub.ID]))
		out[i].StudentName = sub.StudentName
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
		url, _ := h.svc.SignedURL(r.Context(), a.ObjectKey, a.Filename)
		out[i] = AttachmentResponse{ID: a.ID, Filename: a.Filename, MimeType: a.MimeType, SizeBytes: a.SizeBytes, URL: url}
	}
	return out
}

func (h *Handler) buildFiles(r *http.Request, files []*domain.File) []AttachmentResponse {
	if len(files) == 0 {
		return nil
	}
	out := make([]AttachmentResponse, len(files))
	for i, f := range files {
		url, _ := h.svc.SignedURL(r.Context(), f.ObjectKey, f.OriginalName)
		out[i] = AttachmentResponse{ID: f.ID, Filename: f.OriginalName, MimeType: f.MimeType, SizeBytes: f.SizeBytes, URL: url}
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

// writeServiceError — общий маппер: конкретные коды доменных ошибок
// (ALREADY_GRADED, FILE_IN_USE…), конфликты → 409, причина 500 — в лог.
func writeServiceError(w http.ResponseWriter, err error) {
	response.Error(w, err)
}
