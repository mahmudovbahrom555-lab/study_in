package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// DemoGroupCreator is the minimal interface needed to create a group for demo purposes.
type DemoGroupCreator interface {
	CreateDemoGroup(ctx context.Context, teacherID uuid.UUID) (uuid.UUID, error)
}

// Fixed UUIDs for demo data — stable across requests so the frontend can
// navigate to student detail / call recommendation actions without 404s.
// These are synthetic identifiers that never exist in the real tables;
// handler-level guards (isDemoGroup check) must intercept those routes.
var (
	demoStudentIDs = [6]uuid.UUID{
		uuid.MustParse("d1000000-0000-0000-0000-000000000001"),
		uuid.MustParse("d1000000-0000-0000-0000-000000000002"),
		uuid.MustParse("d1000000-0000-0000-0000-000000000003"),
		uuid.MustParse("d1000000-0000-0000-0000-000000000004"),
		uuid.MustParse("d1000000-0000-0000-0000-000000000005"),
		uuid.MustParse("d1000000-0000-0000-0000-000000000006"),
	}
	demoRecIDs = [4]uuid.UUID{
		uuid.MustParse("d2000000-0000-0000-0000-000000000001"),
		uuid.MustParse("d2000000-0000-0000-0000-000000000002"),
		uuid.MustParse("d2000000-0000-0000-0000-000000000003"),
		uuid.MustParse("d2000000-0000-0000-0000-000000000004"),
	}
)

// IsDemoStudentID returns true if the given UUID belongs to the demo student pool.
func IsDemoStudentID(id uuid.UUID) bool {
	for _, sid := range demoStudentIDs {
		if sid == id {
			return true
		}
	}
	return false
}

// IsDemoRecID returns true if the given UUID belongs to the demo recommendation pool.
func IsDemoRecID(id uuid.UUID) bool {
	for _, rid := range demoRecIDs {
		if rid == id {
			return true
		}
	}
	return false
}

// SeedDemoGroup creates a demo group and returns its ID.
// The group is flagged is_demo=true; its AI Insights endpoint returns
// hardcoded rich data so a new teacher sees the full platform value immediately.
func (s *Service) SeedDemoGroup(ctx context.Context, teacherID uuid.UUID) (uuid.UUID, error) {
	if s.demoCreator == nil {
		return uuid.Nil, fmt.Errorf("demo seeder not configured")
	}
	return s.demoCreator.CreateDemoGroup(ctx, teacherID)
}

// DemoInsights returns a fully-populated ClassInsights for a demo group.
// All values are realistic — tuned to show the platform at its most useful.
// Called instead of the DB query when group.is_demo = true.
// Student and recommendation IDs are deterministic (see demoStudentIDs / demoRecIDs)
// so the frontend can safely navigate to detail pages.
func DemoInsights(groupID uuid.UUID) *domain.ClassInsights {
	now := time.Now()

	students := []domain.StudentSummary{
		{StudentID: demoStudentIDs[0], Name: "Алибек Жумаев", XPTotal: 1240, StreakDays: 14,
			TopWeakness: "Passive Voice", IsAtRisk: false, LastActive: strPtr(now.AddDate(0, 0, -1).Format("2006-01-02"))},
		{StudentID: demoStudentIDs[1], Name: "Малика Рашидова", XPTotal: 980, StreakDays: 7,
			TopWeakness: "Conditionals", IsAtRisk: false, LastActive: strPtr(now.AddDate(0, 0, -2).Format("2006-01-02"))},
		{StudentID: demoStudentIDs[2], Name: "Дилноза Каримова", XPTotal: 740, StreakDays: 3,
			TopWeakness: "Past Perfect", IsAtRisk: false, LastActive: strPtr(now.AddDate(0, 0, -1).Format("2006-01-02"))},
		{StudentID: demoStudentIDs[3], Name: "Жасурбек Ибрагимов", XPTotal: 420, StreakDays: 0,
			TopWeakness: "Relative Clauses", IsAtRisk: true, LastActive: strPtr(now.AddDate(0, 0, -5).Format("2006-01-02"))},
		{StudentID: demoStudentIDs[4], Name: "Нилуфар Юсупова", XPTotal: 310, StreakDays: 0,
			TopWeakness: "Passive Voice", IsAtRisk: true, LastActive: strPtr(now.AddDate(0, 0, -4).Format("2006-01-02"))},
		{StudentID: demoStudentIDs[5], Name: "Отабек Хасанов", XPTotal: 1560, StreakDays: 21,
			TopWeakness: "Reported Speech", IsAtRisk: false, LastActive: strPtr(now.Format("2006-01-02"))},
	}

	weakness := []domain.TopicWeakness{
		{Topic: "Passive Voice", AvgAccuracy: 0.38, AvgConfidence: 0.31, AvgConsistency: 0.22,
			StudentsStruggling: 4, TotalStudents: 6},
		{Topic: "Conditionals", AvgAccuracy: 0.51, AvgConfidence: 0.48, AvgConsistency: 0.40,
			StudentsStruggling: 3, TotalStudents: 6},
		{Topic: "Past Perfect", AvgAccuracy: 0.57, AvgConfidence: 0.60, AvgConsistency: 0.55,
			StudentsStruggling: 2, TotalStudents: 6},
		{Topic: "Relative Clauses", AvgAccuracy: 0.62, AvgConfidence: 0.65, AvgConsistency: 0.58,
			StudentsStruggling: 1, TotalStudents: 6},
	}

	recs := []domain.TeacherRecommendation{
		{ID: demoRecIDs[0], Priority: 1, Action: "create_quiz", Topic: "Passive Voice",
			Reason: "Средняя точность 38% и уверенность 31% — 4/6 учеников испытывают трудности", StudentCount: 4},
		{ID: demoRecIDs[1], Priority: 1, Action: "check_students",
			Reason: "2 ученика не заходили 4+ дня — стоит написать им", StudentCount: 2},
		{ID: demoRecIDs[2], Priority: 2, Action: "schedule_review", Topic: "Conditionals",
			Reason: "3 ученика с трудом справляются — назначьте повторение через 3 дня", StudentCount: 3},
		{ID: demoRecIDs[3], Priority: 2, Action: "rate_questions",
			Reason: "TAR 71% — оцените сгенерированные вопросы, чтобы улучшить качество AI"},
	}

	return &domain.ClassInsights{
		GroupID:       groupID,
		Period:        "last_30_days",
		StudentCount:  6,
		ClassWeakness: weakness,
		Students:      students,
		QuizStats: domain.QuizStats{
			Generated:      12,
			AcceptanceRate: 0.71,
			TotalAttempts:  48,
			AvgScore:       67.4,
		},
		Recommendations: recs,
		CourseCoverage: domain.CourseCoverage{
			TopicsCovered:   8,
			TopicsEstimated: 90, // B2 budget
			CoverageRate:    8.0 / 90.0,
		},
		IsDemo: true,
	}
}

func strPtr(s string) *string { return &s }
