package assignments

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// --- Requests ---

type CreateAssignmentRequest struct {
	Title       string     `json:"title"       validate:"required,min=2,max=300"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateAssignmentRequest struct {
	Title       *string    `json:"title"       validate:"omitempty,min=2,max=300"`
	Description *string    `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type SubmitRequest struct {
	Comment *string `json:"comment"`
}

type GradeRequest struct {
	Grade int16   `json:"grade"      validate:"min=0,max=100"`
	Note  *string `json:"note"`
}

// --- Responses ---

type AttachmentResponse struct {
	ID        uuid.UUID `json:"id"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	URL       string    `json:"url"`
}

type AssignmentResponse struct {
	ID          uuid.UUID            `json:"id"`
	GroupID     uuid.UUID            `json:"group_id"`
	TeacherID   uuid.UUID            `json:"teacher_id"`
	Title       string               `json:"title"`
	Description *string              `json:"description"`
	DueDate     *time.Time           `json:"due_date"`
	Attachments []AttachmentResponse `json:"attachments"`
	CreatedAt   time.Time            `json:"created_at"`
}

type SubmissionResponse struct {
	ID           uuid.UUID            `json:"id"`
	AssignmentID uuid.UUID            `json:"assignment_id"`
	StudentID    uuid.UUID            `json:"student_id"`
	Comment      *string              `json:"comment"`
	Grade        *int16               `json:"grade"`
	TeacherNote  *string              `json:"teacher_note"`
	Attachments  []AttachmentResponse `json:"attachments"`
	SubmittedAt  time.Time            `json:"submitted_at"`
	GradedAt     *time.Time           `json:"graded_at"`
}

func assignmentToResponse(a *domain.Assignment, atts []AttachmentResponse) AssignmentResponse {
	return AssignmentResponse{
		ID:          a.ID,
		GroupID:     a.GroupID,
		TeacherID:   a.TeacherID,
		Title:       a.Title,
		Description: a.Description,
		DueDate:     a.DueDate,
		Attachments: atts,
		CreatedAt:   a.CreatedAt,
	}
}

func submissionToResponse(s *domain.Submission, atts []AttachmentResponse) SubmissionResponse {
	return SubmissionResponse{
		ID:           s.ID,
		AssignmentID: s.AssignmentID,
		StudentID:    s.StudentID,
		Comment:      s.Comment,
		Grade:        s.Grade,
		TeacherNote:  s.TeacherNote,
		Attachments:  atts,
		SubmittedAt:  s.SubmittedAt,
		GradedAt:     s.GradedAt,
	}
}
