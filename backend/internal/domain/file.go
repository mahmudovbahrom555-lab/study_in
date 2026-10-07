package domain

import (
	"time"

	"github.com/google/uuid"
)

// FilePurpose — для чего загружен файл. Определяет, кто и куда может его прикрепить.
type FilePurpose string

const FilePurposeSubmission FilePurpose = "submission"

// FileStatus — прикреплён ли файл к сущности. Неприкреплённые удаляются очисткой.
type FileStatus string

const (
	FileStatusPending  FileStatus = "pending"
	FileStatusAttached FileStatus = "attached"
)

// File — загруженный пользователем файл. Единое хранилище для всех разделов:
// разделы ссылаются на File.ID, поэтому любой обработчик (в т.ч. будущий AI)
// получает файл одинаково.
type File struct {
	ID           uuid.UUID   `db:"id"`
	OwnerID      uuid.UUID   `db:"owner_id"`
	Purpose      FilePurpose `db:"purpose"`
	ObjectKey    string      `db:"object_key"`
	OriginalName string      `db:"original_name"`
	MimeType     string      `db:"mime_type"`
	SizeBytes    int64       `db:"size_bytes"`
	SHA256       *string     `db:"sha256"`
	Width        *int        `db:"width"`
	Height       *int        `db:"height"`
	Status       FileStatus  `db:"status"`
	CreatedAt    time.Time   `db:"created_at"`
}
