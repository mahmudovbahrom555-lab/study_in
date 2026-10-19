package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusTrial   PaymentStatus = "trial"
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
)

type Group struct {
	ID          uuid.UUID  `db:"id"`
	TeacherID   uuid.UUID  `db:"teacher_id"`
	Name        string     `db:"name"`
	Subject     *string    `db:"subject"`
	Description *string    `db:"description"`
	InviteCode  string     `db:"invite_code"`
	IsArchived  bool       `db:"is_archived"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	// Добавлена миграцией 000016. Нужна здесь, т.к. репозиторий делает SELECT *,
	// а sqlx падает на колонках без соответствующего поля.
	CEFRLevel *string `db:"cefr_level"`
	// Чат класса (миграция 000026).
	ChatTTLSeconds int      `db:"chat_ttl_seconds"`
	ChatMode       ChatMode `db:"chat_mode"`
}

type GroupMember struct {
	GroupID       uuid.UUID     `db:"group_id"`
	StudentID     uuid.UUID     `db:"student_id"`
	JoinedAt      time.Time     `db:"joined_at"`
	PaymentStatus PaymentStatus `db:"payment_status"`
	ChatMuted     bool          `db:"chat_muted"` // репетитор запретил писать в чат класса
}

// GroupMemberDetail — GroupMember + данные пользователя (для списка участников).
type GroupMemberDetail struct {
	GroupMember
	Name      string  `db:"name"`
	Phone     string  `db:"phone"`
	AvatarURL *string `db:"avatar_url"`
}
