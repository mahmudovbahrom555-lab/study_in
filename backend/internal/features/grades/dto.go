package grades

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type CreateGradeRequest struct {
	StudentID uuid.UUID `json:"student_id" validate:"required"`
	Subject   string    `json:"subject"    validate:"required,max=100"`
	Value     float64   `json:"value"      validate:"required,min=0"`
	MaxValue  float64   `json:"max_value"  validate:"omitempty,min=1"`
	Comment   *string   `json:"comment"`
	GradedAt  *string   `json:"graded_at"`
}

type UpdateGradeRequest struct {
	Subject  *string  `json:"subject"   validate:"omitempty,max=100"`
	Value    *float64 `json:"value"     validate:"omitempty,min=0"`
	MaxValue *float64 `json:"max_value" validate:"omitempty,min=1"`
	Comment  *string  `json:"comment"`
	GradedAt *string  `json:"graded_at"`
}

type GradeResponse struct {
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

func gradeToResponse(g *domain.Grade) GradeResponse {
	return GradeResponse{
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
