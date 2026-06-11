package domain

import (
	"time"

	"github.com/google/uuid"
)

type Grade struct {
	ID        uuid.UUID `db:"id"`
	GroupID   uuid.UUID `db:"group_id"`
	StudentID uuid.UUID `db:"student_id"`
	TeacherID uuid.UUID `db:"teacher_id"`
	Subject   string    `db:"subject"`
	Value     float64   `db:"value"`
	MaxValue  float64   `db:"max_value"`
	Comment   *string   `db:"comment"`
	GradedAt  time.Time `db:"graded_at"`
	CreatedAt time.Time `db:"created_at"`
}
