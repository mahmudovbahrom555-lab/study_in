package domain

import (
	"time"

	"github.com/google/uuid"
)

// ParentReport — monthly ROI snapshot for a student in a group.
type ParentReport struct {
	ID                    uuid.UUID  `db:"id"`
	StudentID             uuid.UUID  `db:"student_id"`
	GroupID               uuid.UUID  `db:"group_id"`
	PeriodStart           time.Time  `db:"period_start"`
	PeriodEnd             time.Time  `db:"period_end"`
	AttendancePct         float64    `db:"attendance_pct"`
	QuizScoreAvg          float64    `db:"quiz_score_avg"`
	QuizScorePrevAvg      float64    `db:"quiz_score_prev_avg"`
	MasteryAvg            float64    `db:"mastery_avg"`
	MasteryPrevAvg        float64    `db:"mastery_prev_avg"`
	HomeworkCompletionPct float64    `db:"homework_completion_pct"`
	QuizAttemptsCount     int        `db:"quiz_attempts_count"`
	AISummary             *string    `db:"ai_summary"`
	CreatedAt             time.Time  `db:"created_at"`
}

// QuizScoreDelta returns percentage-point change in quiz scores.
func (r *ParentReport) QuizScoreDelta() float64 {
	return r.QuizScoreAvg - r.QuizScorePrevAvg
}

// MasteryDelta returns percentage-point change (mastery is 0-1, multiply by 100).
func (r *ParentReport) MasteryDelta() float64 {
	return (r.MasteryAvg - r.MasteryPrevAvg) * 100
}

// RiskAlert — churn-risk signal for a student detected in a teacher's group.
type RiskAlert struct {
	ID            uuid.UUID `db:"id"`
	TeacherID     uuid.UUID `db:"teacher_id"`
	StudentID     uuid.UUID `db:"student_id"`
	GroupID       uuid.UUID `db:"group_id"`
	StudentName   string    `db:"student_name"`
	GroupName     string    `db:"group_name"`
	RiskLevel     string    `db:"risk_level"`
	TriggerReason string    `db:"trigger_reason"`
	Recommendation string   `db:"recommendation"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
}
