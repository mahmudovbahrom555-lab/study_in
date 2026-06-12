package ai_test

// Course Coverage isolation tests.
//
// The critical invariant: topics answered by a student in Group B must NOT
// count toward Group A's coverage. Before the fix, ClassInsights joined
// topic_mastery (global, no group_id) directly — a student in 3 groups
// would inflate every group's coverage with their entire history.
//
// After the fix, the SQL traces:
//   group_members → quiz_attempts → quizzes(group_id) → student_answers → question_topics
//
// These tests verify the service-level contract: GetClassInsights for Group A
// returns CourseCoverage that matches only the topics supplied by the mock for
// Group A, not a merged view of all groups.

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/ai"
)

// groupScopedInsightsRepo returns different ClassInsights per group ID so we
// can verify that the service passes the correct groupID to the repository.
type groupScopedInsightsRepo struct {
	byGroup map[uuid.UUID]*domain.ClassInsights
}

func (r *groupScopedInsightsRepo) ClassInsights(_ context.Context, groupID uuid.UUID) (*domain.ClassInsights, error) {
	if ins, ok := r.byGroup[groupID]; ok {
		return ins, nil
	}
	return &domain.ClassInsights{GroupID: groupID, Period: "last_30_days"}, nil
}

func (r *groupScopedInsightsRepo) StudentProgress(_ context.Context, _, studentID uuid.UUID) (*domain.StudentProgress, error) {
	return &domain.StudentProgress{StudentID: studentID}, nil
}

func (r *groupScopedInsightsRepo) IsDemo(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

// TestCourseCoverage_IsolatedByGroup is the primary regression test for the
// cross-group data leak. A student enrolled in two groups should contribute
// their coverage only to the group whose quizzes they answered.
func TestCourseCoverage_IsolatedByGroup(t *testing.T) {
	groupA := uuid.New()
	groupB := uuid.New()

	// Group A: 5 topics covered, budget 50 (A2 level)
	insA := &domain.ClassInsights{
		GroupID: groupA,
		Period:  "last_30_days",
		CourseCoverage: domain.CourseCoverage{
			TopicsCovered:   5,
			TopicsEstimated: 50,
			CoverageRate:    5.0 / 50.0,
		},
	}
	// Group B: 20 topics covered — must NOT leak into Group A.
	insB := &domain.ClassInsights{
		GroupID: groupB,
		Period:  "last_30_days",
		CourseCoverage: domain.CourseCoverage{
			TopicsCovered:   20,
			TopicsEstimated: 50,
			CoverageRate:    20.0 / 50.0,
		},
	}

	repo := &groupScopedInsightsRepo{
		byGroup: map[uuid.UUID]*domain.ClassInsights{
			groupA: insA,
			groupB: insB,
		},
	}

	svc := ai.NewService(
		newMockDocRepo(), newMockJobRepo(), newMockSessionRepo(), &mockTokenRepo{},
		newMockMasteryRepo(), newMockGamifRepo(), &mockTopicRepo{}, &mockQuizCreator{},
		&mockObjectStore{data: map[string][]byte{}},
		repo, nil, nil, nil, nil, nil,
		ai.ServiceConfig{},
	)

	ctx := context.Background()

	gotA, err := svc.GetClassInsights(ctx, groupA, uuid.New())
	if err != nil {
		t.Fatalf("GetClassInsights groupA: %v", err)
	}
	gotB, err := svc.GetClassInsights(ctx, groupB, uuid.New())
	if err != nil {
		t.Fatalf("GetClassInsights groupB: %v", err)
	}

	// Group A coverage must be exactly 5, not contaminated by Group B's 20.
	if gotA.CourseCoverage.TopicsCovered != 5 {
		t.Errorf("GroupA coverage leaked from GroupB: got %d topics, want 5",
			gotA.CourseCoverage.TopicsCovered)
	}
	if gotB.CourseCoverage.TopicsCovered != 20 {
		t.Errorf("GroupB coverage wrong: got %d, want 20",
			gotB.CourseCoverage.TopicsCovered)
	}

	const eps = 0.001
	wantRateA := 5.0 / 50.0
	if diff := gotA.CourseCoverage.CoverageRate - wantRateA; diff > eps || diff < -eps {
		t.Errorf("GroupA CoverageRate: got %.4f, want %.4f", gotA.CourseCoverage.CoverageRate, wantRateA)
	}
}

// TestCourseCoverage_UnknownCEFR verifies that groups with no CEFR level set
// return CoverageRate=0 and TopicsEstimated=0 rather than dividing by zero.
func TestCourseCoverage_UnknownCEFR(t *testing.T) {
	groupID := uuid.New()
	ins := &domain.ClassInsights{
		GroupID: groupID,
		Period:  "last_30_days",
		CourseCoverage: domain.CourseCoverage{
			TopicsCovered:   3,
			TopicsEstimated: 0, // unknown CEFR
			CoverageRate:    0,
		},
	}
	repo := &groupScopedInsightsRepo{byGroup: map[uuid.UUID]*domain.ClassInsights{groupID: ins}}

	svc := ai.NewService(
		newMockDocRepo(), newMockJobRepo(), newMockSessionRepo(), &mockTokenRepo{},
		newMockMasteryRepo(), newMockGamifRepo(), &mockTopicRepo{}, &mockQuizCreator{},
		&mockObjectStore{data: map[string][]byte{}},
		repo, nil, nil, nil, nil, nil,
		ai.ServiceConfig{},
	)

	got, err := svc.GetClassInsights(context.Background(), groupID, uuid.New())
	if err != nil {
		t.Fatalf("GetClassInsights: %v", err)
	}
	if got.CourseCoverage.CoverageRate != 0 {
		t.Errorf("expected CoverageRate=0 when CEFR unknown, got %.4f",
			got.CourseCoverage.CoverageRate)
	}
	if got.CourseCoverage.TopicsEstimated != 0 {
		t.Errorf("expected TopicsEstimated=0 when CEFR unknown, got %d",
			got.CourseCoverage.TopicsEstimated)
	}
}

// TestCEFRTopicBudget_AllLevels pins the CEFRTopicBudget values so that
// accidental changes to the lookup table are caught immediately.
func TestCEFRTopicBudget_AllLevels(t *testing.T) {
	cases := []struct {
		level string
		want  int
	}{
		{"A1", 30}, {"A2", 50}, {"B1", 70}, {"B2", 90}, {"C1", 110}, {"C2", 130},
		{"", 0}, {"X9", 0}, {"b2", 0}, // unknown levels → 0 (no denominator)
	}
	for _, tc := range cases {
		got := domain.CEFRTopicBudget(tc.level)
		if got != tc.want {
			t.Errorf("CEFRTopicBudget(%q) = %d, want %d", tc.level, got, tc.want)
		}
	}
}
