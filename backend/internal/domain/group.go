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
	IsPublic    bool       `db:"is_public"` // открытая группа видна в поиске
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
	Name         string  `db:"name"`
	Phone        *string `db:"phone"`
	AvatarURL    *string `db:"avatar_url"`
	AvatarPreset *string `db:"avatar_preset"`
}

// JoinRequestStatus — состояние заявки ученика в группу.
type JoinRequestStatus string

const (
	// JoinRequestParentPending — заявка ребёнка младше 13 ждёт согласия родителя.
	JoinRequestParentPending JoinRequestStatus = "parent_pending"
	JoinRequestPending       JoinRequestStatus = "pending"
	JoinRequestAccepted      JoinRequestStatus = "accepted"
	JoinRequestDeclined      JoinRequestStatus = "declined"
)

// GroupJoinRequest — заявка ученика; репетитор принимает или отклоняет.
type GroupJoinRequest struct {
	ID        uuid.UUID         `db:"id"`
	GroupID   uuid.UUID         `db:"group_id"`
	StudentID uuid.UUID         `db:"student_id"`
	Status    JoinRequestStatus `db:"status"`
	CreatedAt time.Time         `db:"created_at"`
	DecidedAt *time.Time        `db:"decided_at"`
	// Родитель, разрешивший заявку ребёнка (миграция 000032).
	ApprovedByParent *uuid.UUID `db:"approved_by_parent"`
}

// JoinRequestDetail — заявка с данными ученика (для репетитора).
type JoinRequestDetail struct {
	GroupJoinRequest
	StudentName  string  `db:"student_name"`
	AvatarURL    *string `db:"avatar_url"`
	AvatarPreset *string `db:"avatar_preset"`
	// Имя родителя, разрешившего заявку ребёнка; nil у учеников 13+.
	ApprovedByParentName *string `db:"approved_by_parent_name"`
}

// MyJoinRequest — своя заявка ученика с названием группы.
type MyJoinRequest struct {
	GroupJoinRequest
	GroupName string `db:"group_name"`
}

// Applicant — кто подаёт заявку: имя (для уведомления родителю) и признак
// ребёнка младше 13.
type Applicant struct {
	Name    string `db:"name"`
	IsChild bool   `db:"is_child"`
}

// ParentRef — привязанный родитель ученика и язык его уведомлений.
type ParentRef struct {
	ID       uuid.UUID `db:"id"`
	Language Language  `db:"language"`
}

// ParentConsent — заявка ребёнка, ждущая согласия родителя.
type ParentConsent struct {
	ID          uuid.UUID `db:"id"`
	StudentID   uuid.UUID `db:"student_id"`
	StudentName string    `db:"student_name"`
	GroupID     uuid.UUID `db:"group_id"`
	GroupName   string    `db:"group_name"`
	Subject     *string   `db:"subject"`
	TeacherName string    `db:"teacher_name"`
	CreatedAt   time.Time `db:"created_at"`
}

// GroupPreview — карточка группы по коду или ссылке: что видит ученик до
// вступления. Без участников и без контактов репетитора.
type GroupPreview struct {
	ID                  uuid.UUID `db:"id"`
	Name                string    `db:"name"`
	Subject             *string   `db:"subject"`
	Description         *string   `db:"description"`
	IsArchived          bool      `db:"is_archived"`
	InviteCode          string    `db:"invite_code"`
	TeacherName         string    `db:"teacher_name"`
	TeacherUsername     *string   `db:"teacher_username"`
	TeacherAvatarURL    *string   `db:"teacher_avatar_url"`
	TeacherAvatarPreset *string   `db:"teacher_avatar_preset"`
	MemberCount         int       `db:"member_count"`
}
