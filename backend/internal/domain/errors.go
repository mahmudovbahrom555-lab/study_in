// Package domain содержит бизнес-сущности и доменные ошибки.
// Этот слой ничего не знает о PostgreSQL, HTTP или других деталях реализации.
package domain

import "errors"

// Доменные ошибки. Используются во всех слоях приложения.
// HTTP слой переводит их в соответствующие коды ответа.
var (
	// ErrNotFound — сущность не найдена.
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized — нет валидной аутентификации.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden — аутентификация есть, но доступ запрещён.
	ErrForbidden = errors.New("forbidden")

	// ErrValidation — данные не прошли валидацию.
	ErrValidation = errors.New("validation failed")

	// ErrConflict — конфликт состояния (например, дубликат).
	ErrConflict = errors.New("conflict")

	// ErrRateLimit — превышен лимит запросов.
	ErrRateLimit = errors.New("rate limit exceeded")

	// ErrInternal — внутренняя ошибка сервера.
	ErrInternal = errors.New("internal server error")

	// ErrAlreadyMember — пользователь уже состоит в группе.
	ErrAlreadyMember = errors.New("already a member")

	// ErrNotMember — пользователь не состоит в группе.
	ErrNotMember = errors.New("not a member")

	// ErrRoleAlreadySet — роль пользователя уже установлена.
	ErrRoleAlreadySet = errors.New("role already set")
)

// Error — типизированная доменная ошибка с дополнительными деталями.
type Error struct {
	Code    string
	Message string
	Cause   error
	Details map[string]any
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError создаёт типизированную доменную ошибку.
func NewError(code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
