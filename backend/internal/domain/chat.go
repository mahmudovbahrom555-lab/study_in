package domain

import (
	"time"

	"github.com/google/uuid"
)

// ChatMode — кто пишет в чат класса.
type ChatMode string

const (
	ChatModeOpen      ChatMode = "open"       // пишут все участники
	ChatModeOwnerOnly ChatMode = "owner_only" // пишет только репетитор (канал)
)

// DefaultChatTTL — срок хранения сообщений нового класса (24 ч).
const DefaultChatTTL = 86400

// ChatTTLs — допустимые сроки хранения сообщений, секунды (1, 6, 12, 24 ч).
// Длинные сроки — платные, этап 3 (docs/PRODUCT-VECTOR.md).
var ChatTTLs = []int{3600, 21600, 43200, DefaultChatTTL}

// ValidChatTTL сообщает, можно ли выбрать такой срок хранения.
func ValidChatTTL(seconds int) bool {
	for _, t := range ChatTTLs {
		if t == seconds {
			return true
		}
	}
	return false
}

// ChatMessage — сообщение в чате класса. Живёт до ExpiresAt.
type ChatMessage struct {
	ID         uuid.UUID `db:"id"`
	GroupID    uuid.UUID `db:"group_id"`
	AuthorID   uuid.UUID `db:"author_id"`
	AuthorName string    `db:"author_name"`
	Body       string    `db:"body"`
	ClientID   uuid.UUID `db:"client_id"`
	CreatedAt  time.Time `db:"created_at"`
	ExpiresAt  time.Time `db:"expires_at"`
}

// ChatAccess — всё, что нужно для проверки прав в чате, одним запросом.
type ChatAccess struct {
	TeacherID  uuid.UUID `db:"teacher_id"`
	IsArchived bool      `db:"is_archived"`
	TTLSeconds int       `db:"chat_ttl_seconds"`
	Mode       ChatMode  `db:"chat_mode"`
	IsMember   bool      `db:"is_member"`
	Muted      bool      `db:"chat_muted"`
	ChatBanned bool      `db:"chat_banned"` // запрет во всех чатах (по жалобе)
	IsActive   bool      `db:"is_active"`
}
