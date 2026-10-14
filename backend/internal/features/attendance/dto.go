package attendance

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const dateLayout = "2006-01-02"

type MarkRequest struct {
	StudentID  uuid.UUID               `json:"student_id"  validate:"required"`
	LessonDate string                  `json:"lesson_date" validate:"required"` // "2006-01-02"
	Status     domain.AttendanceStatus `json:"status"      validate:"required,oneof=present absent late excused"`
	Note       *string                 `json:"note"`
}

// BatchMarkRequest — несколько отметок за один день (кнопка «Остальные были»).
// 100 — с запасом больше любой группы репетитора, защита от огромного запроса.
type BatchMarkRequest struct {
	LessonDate string      `json:"lesson_date" validate:"required"`
	Marks      []BatchMark `json:"marks"       validate:"required,min=1,max=100,dive"`
}

type BatchMark struct {
	StudentID uuid.UUID               `json:"student_id" validate:"required"`
	Status    domain.AttendanceStatus `json:"status"     validate:"required,oneof=present absent late excused"`
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

// RosterEntryResponse — строка списка занятия; attendance_id и status — null,
// если ученик ещё не отмечен.
type RosterEntryResponse struct {
	StudentID    uuid.UUID                `json:"student_id"`
	StudentName  string                   `json:"student_name"`
	AttendanceID *uuid.UUID               `json:"attendance_id"`
	Status       *domain.AttendanceStatus `json:"status"`
	Note         *string                  `json:"note,omitempty"`
}

func toResponse(a *domain.Attendance) AttendanceResponse {
	return AttendanceResponse{
		ID:         a.ID,
		GroupID:    a.GroupID,
		StudentID:  a.StudentID,
		TeacherID:  a.TeacherID,
		LessonDate: a.LessonDate.Format(dateLayout),
		Status:     a.Status,
		Note:       a.Note,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

func toRosterResponse(e *domain.RosterEntry) RosterEntryResponse {
	return RosterEntryResponse{
		StudentID:    e.StudentID,
		StudentName:  e.StudentName,
		AttendanceID: e.AttendanceID,
		Status:       e.Status,
		Note:         e.Note,
	}
}
