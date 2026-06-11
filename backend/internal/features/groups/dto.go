package groups

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// --- Requests ---

type CreateGroupRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=200"`
	Subject     *string `json:"subject"     validate:"omitempty,max=100"`
	Description *string `json:"description"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=200"`
	Subject     *string `json:"subject"     validate:"omitempty,max=100"`
	Description *string `json:"description"`
}

type JoinGroupRequest struct {
	InviteCode string `json:"invite_code" validate:"required,min=4,max=10"`
}

type UpdatePaymentRequest struct {
	Status domain.PaymentStatus `json:"status" validate:"required,oneof=paid pending trial"`
}

// --- Responses ---

type GroupResponse struct {
	ID          uuid.UUID `json:"id"`
	TeacherID   uuid.UUID `json:"teacher_id"`
	Name        string    `json:"name"`
	Subject     *string   `json:"subject"`
	Description *string   `json:"description"`
	InviteCode  string    `json:"invite_code"`
	IsArchived  bool      `json:"is_archived"`
	MemberCount int       `json:"member_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type MemberResponse struct {
	StudentID     uuid.UUID            `json:"student_id"`
	Name          string               `json:"name"`
	Phone         string               `json:"phone"`
	AvatarURL     *string              `json:"avatar_url"`
	JoinedAt      time.Time            `json:"joined_at"`
	PaymentStatus domain.PaymentStatus `json:"payment_status"`
}

func groupToResponse(g *domain.Group) GroupResponse {
	return GroupResponse{
		ID:          g.ID,
		TeacherID:   g.TeacherID,
		Name:        g.Name,
		Subject:     g.Subject,
		Description: g.Description,
		InviteCode:  g.InviteCode,
		IsArchived:  g.IsArchived,
		CreatedAt:   g.CreatedAt,
	}
}

func memberToResponse(m *domain.GroupMemberDetail) MemberResponse {
	return MemberResponse{
		StudentID:     m.StudentID,
		Name:          m.Name,
		Phone:         m.Phone,
		AvatarURL:     m.AvatarURL,
		JoinedAt:      m.JoinedAt,
		PaymentStatus: m.PaymentStatus,
	}
}
