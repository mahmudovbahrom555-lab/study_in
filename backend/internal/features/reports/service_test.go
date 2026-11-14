package reports

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type monthlyReportRepo struct {
	Repository
	from, to time.Time
	report   *domain.ParentReport
}

func (r *monthlyReportRepo) GetLatestReport(context.Context, uuid.UUID, uuid.UUID) (*domain.ParentReport, error) {
	return r.report, nil
}

func (r *monthlyReportRepo) ComputeMetrics(_ context.Context, _, _ uuid.UUID, from, to time.Time) (*RawMetrics, error) {
	r.from, r.to = from, to
	return &RawMetrics{}, nil
}

func (r *monthlyReportRepo) SaveReport(_ context.Context, report *domain.ParentReport) error {
	r.report = report
	return nil
}

func TestMonthlyReportUsesExclusiveEndAndPreservesDisplayedDates(t *testing.T) {
	repo := &monthlyReportRepo{}
	report, err := NewService(repo).GenerateParentReport(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, 1, repo.from.Day())
	require.Equal(t, repo.from.AddDate(0, 1, 0), repo.to)
	require.Equal(t, repo.from, report.PeriodStart)
	require.Equal(t, repo.to.AddDate(0, 0, -1), report.PeriodEnd)
	require.True(t, report.PeriodEnd.Add(23*time.Hour).Before(repo.to))
}
