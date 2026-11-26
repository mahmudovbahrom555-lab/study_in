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
	IsPublic    bool    `json:"is_public"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=200"`
	Subject     *string `json:"subject"     validate:"omitempty,max=100"`
	Description *string `json:"description"`
	IsPublic    *bool   `json:"is_public"`
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
	IsPublic    bool      `json:"is_public"`
	MemberCount int       `json:"member_count,omitempty"`
	// Чат класса (docs/design/CHAT.md): непрочитанные — только в списке своих классов.
	UnreadCount    int       `json:"unread_count"`
	ChatTTLSeconds int       `json:"chat_ttl_seconds"`
	ChatMode       string    `json:"chat_mode"`
	CreatedAt      time.Time `json:"created_at"`
}

type MemberResponse struct {
	StudentID     uuid.UUID            `json:"student_id"`
	Name          string               `json:"name"`
	Phone         string               `json:"phone,omitempty"` // только репетитору группы
	AvatarURL     *string              `json:"avatar_url"`
	AvatarPreset  *string              `json:"avatar_preset"`
	JoinedAt      time.Time            `json:"joined_at"`
	PaymentStatus domain.PaymentStatus `json:"payment_status,omitempty"` // только репетитору группы
	ChatMuted     bool                 `json:"chat_muted,omitempty"`     // только репетитору группы
}

func groupToResponse(g *domain.Group) GroupResponse {
	return GroupResponse{
		ID:             g.ID,
		TeacherID:      g.TeacherID,
		Name:           g.Name,
		Subject:        g.Subject,
		Description:    g.Description,
		InviteCode:     g.InviteCode,
		IsArchived:     g.IsArchived,
		IsPublic:       g.IsPublic,
		ChatTTLSeconds: g.ChatTTLSeconds,
		ChatMode:       string(g.ChatMode),
		CreatedAt:      g.CreatedAt,
	}
}

func memberToResponse(m *domain.GroupMemberDetail) MemberResponse {
	phone := ""
	if m.Phone != nil {
		phone = *m.Phone
	}
	return MemberResponse{
		StudentID:     m.StudentID,
		Name:          m.Name,
		Phone:         phone,
		AvatarPreset:  m.AvatarPreset,
		JoinedAt:      m.JoinedAt,
		PaymentStatus: m.PaymentStatus,
		ChatMuted:     m.ChatMuted,
	}
}

// JoinRequestResponse — заявка ученика (ему же в ответ на «вступить»).
type JoinRequestResponse struct {
	ID        uuid.UUID                `json:"id"`
	GroupID   uuid.UUID                `json:"group_id"`
	GroupName string                   `json:"group_name"`
	Status    domain.JoinRequestStatus `json:"status"`
}

// JoinRequestItem — заявка в списке репетитора.
type JoinRequestItem struct {
	ID           uuid.UUID `json:"id"`
	StudentID    uuid.UUID `json:"student_id"`
	StudentName  string    `json:"student_name"`
	AvatarURL    *string   `json:"avatar_url"`
	AvatarPreset *string   `json:"avatar_preset"`
	CreatedAt    time.Time `json:"created_at"`
}

// GroupPreviewResponse — карточка группы по ссылке, до вступления.
type GroupPreviewResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Subject             *string   `json:"subject"`
	Description         *string   `json:"description"`
	TeacherName         string    `json:"teacher_name"`
	TeacherAvatarURL    *string   `json:"teacher_avatar_url"`
	TeacherAvatarPreset *string   `json:"teacher_avatar_preset"`
	MemberCount         int       `json:"member_count"`
}

// GroupSearchItem — открытая группа в поиске: карточка + код, по которому
// ученик отправляет заявку (сама группа всё равно ждёт одобрения репетитора).
type GroupSearchItem struct {
	GroupPreviewResponse
	InviteCode      string  `json:"invite_code"`
	TeacherUsername *string `json:"teacher_username"`
}
