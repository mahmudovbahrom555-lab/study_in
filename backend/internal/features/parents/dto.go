package parents

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type LinkByCodeRequest struct {
	Code string `json:"code" validate:"required,min=4,max=12"`
}

type LinkCodeResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ParentLinkResponse struct {
	ID          uuid.UUID `json:"id"`
	ParentID    uuid.UUID `json:"parent_id"`
	StudentID   uuid.UUID `json:"student_id"`
	StudentName string    `json:"student_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// ChildGroupResponse — группа ребёнка глазами родителя: без кода приглашения
// (иначе родитель мог бы раздать его) и служебных полей.
type ChildGroupResponse struct {
	ID          uuid.UUID `json:"id"`
	TeacherID   uuid.UUID `json:"teacher_id"`
	Name        string    `json:"name"`
	Subject     *string   `json:"subject"`
	Description *string   `json:"description"`
	IsArchived  bool      `json:"is_archived"`
	CreatedAt   time.Time `json:"created_at"`
}

type ChildGradeResponse struct {
	ID        uuid.UUID `json:"id"`
	GroupID   uuid.UUID `json:"group_id"`
	StudentID uuid.UUID `json:"student_id"`
	TeacherID uuid.UUID `json:"teacher_id"`
	Subject   string    `json:"subject"`
	Value     float64   `json:"value"`
	MaxValue  float64   `json:"max_value"`
	Comment   *string   `json:"comment,omitempty"`
	GradedAt  time.Time `json:"graded_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ChildAttendanceResponse struct {
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

func toChildGroup(g *domain.Group) ChildGroupResponse {
	return ChildGroupResponse{
		ID:          g.ID,
		TeacherID:   g.TeacherID,
		Name:        g.Name,
		Subject:     g.Subject,
		Description: g.Description,
		IsArchived:  g.IsArchived,
		CreatedAt:   g.CreatedAt,
	}
}

func toChildGrade(g *domain.Grade) ChildGradeResponse {
	return ChildGradeResponse{
		ID:        g.ID,
		GroupID:   g.GroupID,
		StudentID: g.StudentID,
		TeacherID: g.TeacherID,
		Subject:   g.Subject,
		Value:     g.Value,
		MaxValue:  g.MaxValue,
		Comment:   g.Comment,
		GradedAt:  g.GradedAt,
		CreatedAt: g.CreatedAt,
	}
}

func toChildAttendance(a *domain.Attendance) ChildAttendanceResponse {
	return ChildAttendanceResponse{
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

// mapAll — список ответов из списка доменных объектов.
func mapAll[T any, R any](in []*T, f func(*T) R) []R {
	out := make([]R, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}
