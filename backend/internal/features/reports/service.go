package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Service implements Parent ROI reporting and Owner churn-risk detection.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GenerateParentReport computes or returns the cached monthly report for a student in a group.
// Parent must be the caller; access check (parentID ↔ studentID) is done at handler level.
func (s *Service) GenerateParentReport(ctx context.Context, studentID, groupID uuid.UUID) (*domain.ParentReport, error) {
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)

	// Return cached if it exists and was generated today.
	existing, err := s.repo.GetLatestReport(ctx, studentID, groupID)
	if err == nil && existing != nil {
		today := now.Truncate(24 * time.Hour)
		if !existing.CreatedAt.Before(today) {
			return existing, nil
		}
	}

	metrics, err := s.repo.ComputeMetrics(ctx, studentID, groupID, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("reports.GenerateParentReport compute: %w", err)
	}

	report := &domain.ParentReport{
		StudentID:             studentID,
		GroupID:               groupID,
		PeriodStart:           periodStart,
		PeriodEnd:             periodEnd,
		AttendancePct:         metrics.AttendancePct,
		QuizScoreAvg:          metrics.QuizScoreAvg,
		QuizScorePrevAvg:      metrics.QuizScorePrevAvg,
		MasteryAvg:            metrics.MasteryAvg,
		MasteryPrevAvg:        metrics.MasteryPrevAvg,
		HomeworkCompletionPct: metrics.HomeworkCompletionPct,
		QuizAttemptsCount:     metrics.QuizAttemptsCount,
	}

	if err := s.repo.SaveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("reports.GenerateParentReport save: %w", err)
	}

	// Re-fetch to get the DB-assigned ID and created_at.
	saved, err := s.repo.GetLatestReport(ctx, studentID, groupID)
	if err != nil {
		return report, nil
	}
	return saved, nil
}

// RefreshRiskAlerts recalculates risk signals and persists them for a teacher.
func (s *Service) RefreshRiskAlerts(ctx context.Context, teacherID uuid.UUID) ([]*domain.RiskAlert, error) {
	alerts, err := s.repo.ComputeRiskAlerts(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("reports.RefreshRiskAlerts compute: %w", err)
	}
	if len(alerts) > 0 {
		if err := s.repo.SaveRiskAlerts(ctx, alerts); err != nil {
			return nil, fmt.Errorf("reports.RefreshRiskAlerts save: %w", err)
		}
	}
	return alerts, nil
}

// GetRiskAlerts returns risk alerts for a teacher, optionally filtered by status.
func (s *Service) GetRiskAlerts(ctx context.Context, teacherID uuid.UUID, status string) ([]*domain.RiskAlert, error) {
	return s.repo.GetRiskAlerts(ctx, teacherID, status)
}

// ResolveRiskAlert marks an alert as resolved.
func (s *Service) ResolveRiskAlert(ctx context.Context, alertID, teacherID uuid.UUID) error {
	return s.repo.ResolveRiskAlert(ctx, alertID, teacherID)
}
