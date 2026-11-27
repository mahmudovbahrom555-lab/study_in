package parents

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Привязка родителя по коду, который выдаёт ученик (бриф: «родитель
// привязывается к ребёнку по коду»). Код короткий, без похожих символов
// (0/O, 1/I), чтобы его легко продиктовать или переслать.
const (
	linkCodeLen      = 6
	linkCodeTTL      = 24 * time.Hour
	linkCodeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	linkCodeAttempts = 5

	// Ссылка-приглашение родителю (ученик пересылает её в мессенджере):
	// родитель может зарегистрироваться не сразу, поэтому код живёт неделю,
	// а чтобы его было не подобрать за это время — он длиннее.
	inviteCodeLen = 8
	inviteCodeTTL = 7 * 24 * time.Hour
)

// CreateLinkCode выдаёт ученику новый код для привязки родителя.
// Код многоразовый в течение 24 часов — его можно дать обоим родителям.
// invite — код для ссылки-приглашения: 8 символов, 7 дней.
func (s *Service) CreateLinkCode(ctx context.Context, studentID uuid.UUID, invite bool) (*domain.ParentLinkCode, error) {
	length, ttl := linkCodeLen, linkCodeTTL
	if invite {
		length, ttl = inviteCodeLen, inviteCodeTTL
	}
	now := time.Now()
	for i := 0; i < linkCodeAttempts; i++ {
		code, err := generateLinkCode(length)
		if err != nil {
			return nil, fmt.Errorf("parents.CreateLinkCode generate: %w", err)
		}
		// Код уже занят другим активным кодом — пробуем ещё раз.
		if existing, err := s.repo.GetActiveLinkCode(ctx, code); err != nil {
			return nil, fmt.Errorf("parents.CreateLinkCode check: %w", err)
		} else if existing != nil {
			continue
		}
		c := &domain.ParentLinkCode{
			Code:      code,
			StudentID: studentID,
			ExpiresAt: now.Add(ttl),
			CreatedAt: now,
		}
		if err := s.repo.CreateLinkCode(ctx, c); err != nil {
			return nil, fmt.Errorf("parents.CreateLinkCode: %w", err)
		}
		return c, nil
	}
	return nil, fmt.Errorf("parents.CreateLinkCode: no free code after %d attempts", linkCodeAttempts)
}

// LinkByCode привязывает родителя к ученику, выдавшему код.
func (s *Service) LinkByCode(ctx context.Context, parentID uuid.UUID, code string) (*domain.ParentLink, error) {
	c, err := s.repo.GetActiveLinkCode(ctx, normalizeLinkCode(code))
	if err != nil {
		return nil, fmt.Errorf("parents.LinkByCode: %w", err)
	}
	if c == nil {
		return nil, domain.NewError("CODE_INVALID", "Код неверный или истёк", domain.ErrValidation)
	}
	return s.LinkChild(ctx, parentID, c.StudentID)
}

func normalizeLinkCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), " ", ""))
}

func generateLinkCode(length int) (string, error) {
	b := make([]byte, length)
	n := big.NewInt(int64(len(linkCodeAlphabet)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		b[i] = linkCodeAlphabet[idx.Int64()]
	}
	return string(b), nil
}
