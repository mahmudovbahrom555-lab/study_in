// Package google проверяет ID-токены «Войти через Google».
package google

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/idtoken"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/auth"
)

// Verifier проверяет подпись (ключи Google), срок и audience ID-токена.
type Verifier struct {
	clientID string
}

// NewVerifier — clientID: OAuth-клиент типа «Web application». Android-приложение
// запрашивает токен с ним же (serverClientId), поэтому audience один.
func NewVerifier(clientID string) *Verifier {
	return &Verifier{clientID: clientID}
}

func (v *Verifier) Verify(ctx context.Context, token string) (*auth.GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, token, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("validate google id token: %w", err)
	}
	if payload.Subject == "" {
		return nil, errors.New("google id token without sub")
	}
	id := &auth.GoogleIdentity{Sub: payload.Subject}
	// Почту берём только подтверждённую Google.
	if verified, _ := payload.Claims["email_verified"].(bool); verified {
		id.Email, _ = payload.Claims["email"].(string)
	}
	id.Name, _ = payload.Claims["name"].(string)
	return id, nil
}
