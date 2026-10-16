package domain

import (
	"time"

	"github.com/google/uuid"
)

// GradeKind — откуда оценка в журнале.
type GradeKind string

const (
	// GradeKindClass — репетитор поставил в журнале (работа на занятии).
	GradeKindClass GradeKind = "class"
	// GradeKindHomework — оценка за ДЗ: ставится при проверке работы,
	// меняется только там же (журнал её не правит и не удаляет).
	GradeKindHomework GradeKind = "homework"
)

type Grade struct {
	ID           uuid.UUID  `db:"id"`
	GroupID      uuid.UUID  `db:"group_id"`
	StudentID    uuid.UUID  `db:"student_id"`
	TeacherID    uuid.UUID  `db:"teacher_id"`
	Subject      string     `db:"subject"`
	Value        float64    `db:"value"`
	MaxValue     float64    `db:"max_value"`
	Comment      *string    `db:"comment"`
	Kind         GradeKind  `db:"kind"`
	AssignmentID *uuid.UUID `db:"assignment_id"`
	GradedAt     time.Time  `db:"graded_at"`
	CreatedAt    time.Time  `db:"created_at"`
	// StudentName заполняется в списках журнала (JOIN users).
	StudentName string `db:"student_name"`
}
