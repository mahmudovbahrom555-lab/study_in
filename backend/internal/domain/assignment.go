package domain

import (
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID          uuid.UUID  `db:"id"`
	GroupID     uuid.UUID  `db:"group_id"`
	TeacherID   uuid.UUID  `db:"teacher_id"`
	Title       string     `db:"title"`
	Description *string    `db:"description"`
	DueDate     *time.Time `db:"due_date"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

type AssignmentAttachment struct {
	ID           uuid.UUID `db:"id"`
	AssignmentID uuid.UUID `db:"assignment_id"`
	ObjectKey    string    `db:"object_key"`
	Filename     string    `db:"filename"`
	MimeType     string    `db:"mime_type"`
	SizeBytes    int64     `db:"size_bytes"`
	CreatedAt    time.Time `db:"created_at"`
}

type Submission struct {
	ID           uuid.UUID  `db:"id"`
	AssignmentID uuid.UUID  `db:"assignment_id"`
	StudentID    uuid.UUID  `db:"student_id"`
	Comment      *string    `db:"comment"`
	Grade        *int16     `db:"grade"`
	TeacherNote  *string    `db:"teacher_note"`
	SubmittedAt  time.Time  `db:"submitted_at"`
	GradedAt     *time.Time `db:"graded_at"`
}

type SubmissionAttachment struct {
	ID           uuid.UUID `db:"id"`
	SubmissionID uuid.UUID `db:"submission_id"`
	ObjectKey    string    `db:"object_key"`
	Filename     string    `db:"filename"`
	MimeType     string    `db:"mime_type"`
	SizeBytes    int64     `db:"size_bytes"`
	CreatedAt    time.Time `db:"created_at"`
}

// AssignmentWithMeta — задание с дополнительными полями для ответа API.
type AssignmentWithMeta struct {
	Assignment
	AttachmentCount int        `db:"attachment_count"`
	Submission      *Submission `db:"-"` // nil если студент ещё не сдал
}
