package domain

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "present"
	AttendanceAbsent  AttendanceStatus = "absent"
	AttendanceLate    AttendanceStatus = "late"
	AttendanceExcused AttendanceStatus = "excused"
)

type Attendance struct {
	ID         uuid.UUID        `db:"id"`
	GroupID    uuid.UUID        `db:"group_id"`
	StudentID  uuid.UUID        `db:"student_id"`
	TeacherID  uuid.UUID        `db:"teacher_id"`
	LessonDate time.Time        `db:"lesson_date"`
	Status     AttendanceStatus `db:"status"`
	Note       *string          `db:"note"`
	CreatedAt  time.Time        `db:"created_at"`
	UpdatedAt  time.Time        `db:"updated_at"`
}

// RosterEntry — ученик в списке занятия и его отметка за день
// (поля отметки пустые, если ещё не отмечен).
type RosterEntry struct {
	StudentID    uuid.UUID         `db:"student_id"`
	StudentName  string            `db:"student_name"`
	AttendanceID *uuid.UUID        `db:"attendance_id"`
	Status       *AttendanceStatus `db:"status"`
	Note         *string           `db:"note"`
}
