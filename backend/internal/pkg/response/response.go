// Package response стандартизирует формат HTTP ответов API.
//
// Успех:
//
//	{ "data": ..., "meta": ... }
//
// Ошибка:
//
//	{ "error": { "code": "...", "message": "...", "details": {...} } }
package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// SuccessResponse — стандартный успешный ответ.
type SuccessResponse struct {
	Data any            `json:"data"`
	Meta map[string]any `json:"meta,omitempty"`
}

// ErrorResponse — стандартный ответ с ошибкой.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// OK отправляет успешный ответ со статусом 200.
func OK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// Created отправляет ответ со статусом 201.
func Created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, SuccessResponse{Data: data})
}

// NoContent отправляет пустой ответ со статусом 204.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// WithMeta отправляет успешный ответ с метаданными (например, пагинация).
func WithMeta(w http.ResponseWriter, data any, meta map[string]any) {
	writeJSON(w, http.StatusOK, SuccessResponse{Data: data, Meta: meta})
}

// Error отправляет ответ с ошибкой. Автоматически определяет HTTP статус
// по типу доменной ошибки.
func Error(w http.ResponseWriter, err error) {
	status, code, message := classifyError(err)

	body := ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	}

	var domainErr *domain.Error
	if errors.As(err, &domainErr) && len(domainErr.Details) > 0 {
		body.Error.Details = domainErr.Details
	}

	writeJSON(w, status, body)
}

// ErrorWithDetails отправляет ошибку с дополнительными деталями (например, поля валидации).
func ErrorWithDetails(w http.ResponseWriter, err error, details map[string]any) {
	status, code, message := classifyError(err)
	writeJSON(w, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// classifyError определяет HTTP статус и код по типу доменной ошибки.
func classifyError(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "Resource not found"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", "Access denied"
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR", err.Error()
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "CONFLICT", err.Error()
	case errors.Is(err, domain.ErrRateLimit):
		return http.StatusTooManyRequests, "RATE_LIMIT", "Too many requests"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
