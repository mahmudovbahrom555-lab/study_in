package grades

import (
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type CreateGradeRequest struct {
	StudentID uuid.UUID `json:"student_id" validate:"required"`
	Subject   string    `json:"subject"    validate:"required,max=100"`
	// Указатель: 0 — законная оценка, а required для числа 0 не пропускает.
	Value    *float64 `json:"value"     validate:"required,min=0"`
	MaxValue float64  `json:"max_value" validate:"omitempty,gt=0"`
	Comment  *string  `json:"comment"`
	GradedAt *string  `json:"graded_at"`
}

type UpdateGradeRequest struct {
	Subject  *string  `json:"subject"   validate:"omitempty,max=100"`
	Value    *float64 `json:"value"     validate:"omitempty,min=0"`
	MaxValue *float64 `json:"max_value" validate:"omitempty,gt=0"`
	Comment  *string  `json:"comment"`
	GradedAt *string  `json:"graded_at"`
}

type GradeResponse struct {
	ID           uuid.UUID  `json:"id"`
	GroupID      uuid.UUID  `json:"group_id"`
	StudentID    uuid.UUID  `json:"student_id"`
	StudentName  string     `json:"student_name"`
	TeacherID    uuid.UUID  `json:"teacher_id"`
	Subject      string     `json:"subject"`
	Value        float64    `json:"value"`
	MaxValue     float64    `json:"max_value"`
	Comment      *string    `json:"comment,omitempty"`
	Kind         string     `json:"kind"`
	AssignmentID *uuid.UUID `json:"assignment_id,omitempty"`
	GradedAt     time.Time  `json:"graded_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

func gradeToResponse(g *domain.Grade) GradeResponse {
	return GradeResponse{
		ID:           g.ID,
		GroupID:      g.GroupID,
		StudentID:    g.StudentID,
		StudentName:  g.StudentName,
		TeacherID:    g.TeacherID,
		Subject:      g.Subject,
		Value:        g.Value,
		MaxValue:     g.MaxValue,
		Comment:      g.Comment,
		Kind:         string(g.Kind),
		AssignmentID: g.AssignmentID,
		GradedAt:     g.GradedAt,
		CreatedAt:    g.CreatedAt,
	}
}
