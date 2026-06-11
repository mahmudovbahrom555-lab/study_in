package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/jwt"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/response"
)

type contextKey string

const (
	ctxUserID contextKey = "user_id"
	ctxRole   contextKey = "role"
)

// Auth — middleware, которая проверяет Bearer JWT и кладёт userID/role в контекст.
func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.Error(w, domain.ErrUnauthorized)
				return
			}
			claims, err := jwtManager.ParseAccess(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				response.Error(w, domain.ErrUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromCtx извлекает UUID пользователя из контекста запроса.
func UserIDFromCtx(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(ctxUserID).(uuid.UUID)
	return id
}

// RoleFromCtx извлекает роль пользователя из контекста запроса.
func RoleFromCtx(ctx context.Context) domain.Role {
	role, _ := ctx.Value(ctxRole).(domain.Role)
	return role
}
