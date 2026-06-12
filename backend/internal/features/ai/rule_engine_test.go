package ai_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	ai "github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/ai"
)

func baseContext(students int) ai.RuleContext {
	ins := &domain.ClassInsights{
		GroupID:      uuid.New(),
		StudentCount: students,
		QuizStats:    domain.QuizStats{Generated: 5, AcceptanceRate: 0.80, AvgScore: 60},
	}
	for i := 0; i < students; i++ {
		ins.Students = append(ins.Students, domain.StudentSummary{
			StudentID: uuid.New(),
			IsAtRisk:  false,
		})
	}
	return ai.RuleContext{
		GroupID:   ins.GroupID,
		TeacherID: uuid.New(),
		Insights:  ins,
		TAR:       0.80,
		TARTotal:  20,
	}
}

func TestRuleEngine_MasteryCritical_EmitsCreateQuiz_P1(t *testing.T) {
	rc := baseContext(6)
	rc.Insights.ClassWeakness = []domain.TopicWeakness{
		{Topic: "Past Perfect", AvgAccuracy: 0.35, AvgConfidence: 0.30, AvgConsistency: 0.50,
			StudentsStruggling: 2, TotalStudents: 6},
	}
	recs := ai.RunRules(rc)
	found := false
	for _, r := range recs {
		if r.RuleKey == "mastery_critical" && r.Priority == 1 && r.Action == "create_quiz" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mastery_critical P1 create_quiz, got %+v", recs)
	}
}

func TestRuleEngine_MasteryLow_HalfStruggling_EmitsP1(t *testing.T) {
	rc := baseContext(6)
	rc.Insights.ClassWeakness = []domain.TopicWeakness{
		{Topic: "Conditionals", AvgAccuracy: 0.48, AvgConfidence: 0.55, AvgConsistency: 0.60,
			StudentsStruggling: 3, TotalStudents: 6},
	}
	recs := ai.RunRules(rc)
	found := false
	for _, r := range recs {
		if r.RuleKey == "mastery_low" && r.Priority == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected mastery_low P1, got %+v", recs)
	}
}

func TestRuleEngine_ConsistencyCritical_EmitsScheduleReview(t *testing.T) {
	rc := baseContext(4)
	rc.Insights.ClassWeakness = []domain.TopicWeakness{
		{Topic: "Articles", AvgAccuracy: 0.55, AvgConfidence: 0.60, AvgConsistency: 0.20,
			StudentsStruggling: 1, TotalStudents: 4},
	}
	recs := ai.RunRules(rc)
	found := false
	for _, r := range recs {
		if r.RuleKey == "consistency_critical" && r.Action == "schedule_review" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected consistency_critical schedule_review, got %+v", recs)
	}
}

func TestRuleEngine_AtRisk_EmitsCheckStudents(t *testing.T) {
	rc := baseContext(4)
	rc.Insights.Students[0].IsAtRisk = true
	rc.Insights.Students[1].IsAtRisk = true
	recs := ai.RunRules(rc)
	found := false
	for _, r := range recs {
		if r.RuleKey == "at_risk" && r.StudentCount == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected at_risk with student_count=2, got %+v", recs)
	}
}

func TestRuleEngine_TARLow_EmitsRateQuestions(t *testing.T) {
	rc := baseContext(4)
	rc.TAR = 0.55
	rc.TARTotal = 20
	recs := ai.RunRules(rc)
	found := false
	for _, r := range recs {
		if r.RuleKey == "tar_low" && r.Action == "rate_questions" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected tar_low rate_questions, got %+v", recs)
	}
}

func TestRuleEngine_TARLow_IgnoredWhenInsufficientData(t *testing.T) {
	rc := baseContext(4)
	rc.TAR = 0.55
	rc.TARTotal = 3 // < 5 threshold
	recs := ai.RunRules(rc)
	for _, r := range recs {
		if r.RuleKey == "tar_low" {
			t.Fatalf("tar_low should not fire with only %d rated questions", rc.TARTotal)
		}
	}
}

func TestRuleEngine_Celebrate_NoAtRisk_HighScore(t *testing.T) {
	rc := baseContext(5)
	rc.Insights.QuizStats.AvgScore = 90
	recs := ai.RunRules(rc)
	found := false
	for _, r := range recs {
		if r.RuleKey == "celebrate" && r.Priority == 3 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected celebrate P3, got %+v", recs)
	}
}

func TestRuleEngine_Celebrate_SuppressedWhenAtRisk(t *testing.T) {
	rc := baseContext(5)
	rc.Insights.QuizStats.AvgScore = 90
	rc.Insights.Students[0].IsAtRisk = true
	recs := ai.RunRules(rc)
	for _, r := range recs {
		if r.RuleKey == "celebrate" {
			t.Fatal("celebrate should not fire when students are at-risk")
		}
	}
}

func TestRuleEngine_EachRecHasID(t *testing.T) {
	rc := baseContext(6)
	rc.Insights.ClassWeakness = []domain.TopicWeakness{
		{Topic: "Grammar", AvgAccuracy: 0.40, AvgConfidence: 0.50, AvgConsistency: 0.50,
			StudentsStruggling: 3, TotalStudents: 6},
	}
	recs := ai.RunRules(rc)
	for _, r := range recs {
		if r.ID == uuid.Nil {
			t.Fatalf("recommendation %s has nil ID", r.RuleKey)
		}
	}
}
