package attendance

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type MarkRequest struct {
	StudentID  uuid.UUID               `json:"student_id"  validate:"required"`
	LessonDate string                  `json:"lesson_date" validate:"required"` // "2006-01-02"
	Status     domain.AttendanceStatus `json:"status"      validate:"required,oneof=present absent late excused"`
	Note       *string                 `json:"note"`
}

type UpdateRequest struct {
	Status domain.AttendanceStatus `json:"status" validate:"required,oneof=present absent late excused"`
	Note   *string                 `json:"note"`
}

type AttendanceResponse struct {
	ID         uuid.UUID               `json:"id"`
	GroupID    uuid.UUID               `json:"group_id"`
	StudentID  uuid.UUID               `json:"student_id"`
	TeacherID  uuid.UUID               `json:"teacher_id"`
	LessonDate string                  `json:"lesson_date"`
	Status     domain.AttendanceStatus `json:"status"`
	Note       *string                 `json:"note,omitempty"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}

func toResponse(a *domain.Attendance) AttendanceResponse {
	return AttendanceResponse{
		ID:         a.ID,
		GroupID:    a.GroupID,
		StudentID:  a.StudentID,
		TeacherID:  a.TeacherID,
		LessonDate: a.LessonDate.Format("2006-01-02"),
		Status:     a.Status,
		Note:       a.Note,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
