package reports

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository abstracts all data access for the reports feature.
type Repository interface {
	// ComputeMetrics computes raw metrics for a student in a group for [from, to].
	ComputeMetrics(ctx context.Context, studentID, groupID uuid.UUID, from, to time.Time) (*RawMetrics, error)
	// SaveReport persists a generated report (upserts by student+group+period_start).
	SaveReport(ctx context.Context, r *domain.ParentReport) error
	// GetLatestReport returns the most recent report for student+group.
	GetLatestReport(ctx context.Context, studentID, groupID uuid.UUID) (*domain.ParentReport, error)
	// ComputeRiskAlerts runs SQL heuristics and returns at-risk students for a teacher.
	ComputeRiskAlerts(ctx context.Context, teacherID uuid.UUID) ([]*domain.RiskAlert, error)
	// SaveRiskAlerts replaces today's alerts for the teacher (idempotent refresh).
	SaveRiskAlerts(ctx context.Context, alerts []*domain.RiskAlert) error
	// GetRiskAlerts returns saved alerts for a teacher filtered by status.
	GetRiskAlerts(ctx context.Context, teacherID uuid.UUID, status string) ([]*domain.RiskAlert, error)
	// ResolveRiskAlert marks a single alert as resolved.
	ResolveRiskAlert(ctx context.Context, alertID, teacherID uuid.UUID) error
	// IsGroupMember verifies the student belongs to the group.
	IsGroupMember(ctx context.Context, groupID, studentID uuid.UUID) (bool, error)
	// IsGroupTeacher verifies the caller owns the group.
	IsGroupTeacher(ctx context.Context, groupID, teacherID uuid.UUID) (bool, error)
}

// RawMetrics contains computed statistics for a period.
type RawMetrics struct {
	AttendancePct         float64
	QuizScoreAvg          float64
	QuizScorePrevAvg      float64
	MasteryAvg            float64
	MasteryPrevAvg        float64
	HomeworkCompletionPct float64
	QuizAttemptsCount     int
}
