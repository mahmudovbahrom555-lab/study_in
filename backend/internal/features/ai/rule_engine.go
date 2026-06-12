package ai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// RuleContext is the snapshot of class data evaluated by the Rule Engine.
type RuleContext struct {
	GroupID   uuid.UUID
	TeacherID uuid.UUID
	Insights  *domain.ClassInsights // raw insights (no recommendations yet)
	TAR       float64               // teacher acceptance rate [0,1]
	TARTotal  int                   // total rated questions (TAR is unreliable when small)
}

// RunRules evaluates all rules deterministically and returns a list of recommendations.
// No GPT calls — all logic is pure Go.
// Rules are ordered by impact; topics are processed once to avoid duplicates.
func RunRules(rc RuleContext) []*domain.AIRecommendation {
	var recs []*domain.AIRecommendation
	ins := rc.Insights

	atRisk := 0
	for _, s := range ins.Students {
		if s.IsAtRisk {
			atRisk++
		}
	}

	// Process topic-based rules.
	seenTopic := make(map[string]bool)
	for _, w := range ins.ClassWeakness {
		if seenTopic[w.Topic] {
			continue
		}
		seenTopic[w.Topic] = true

		half := w.TotalStudents / 2
		if half < 1 {
			half = 1
		}

		// Rule mastery_critical: low accuracy + low confidence → create quiz is urgent.
		if w.AvgAccuracy < 0.40 && w.AvgConfidence < 0.35 && w.StudentsStruggling > 0 {
			recs = append(recs, rec(rc, 1, "create_quiz", w.Topic,
				fmt.Sprintf("Средняя точность %.0f%% и уверенность %.0f%% — %d учеников не понимают тему",
					w.AvgAccuracy*100, w.AvgConfidence*100, w.StudentsStruggling),
				w.StudentsStruggling, "mastery_critical", w))
			continue
		}

		// Rule mastery_low: ≥50% of students struggling → create quiz.
		if w.StudentsStruggling >= half {
			recs = append(recs, rec(rc, 1, "create_quiz", w.Topic,
				fmt.Sprintf("%.0f%% средняя точность — %d/%d учеников испытывают трудности",
					w.AvgAccuracy*100, w.StudentsStruggling, w.TotalStudents),
				w.StudentsStruggling, "mastery_low", w))
			continue
		}

		// Rule consistency_critical: consistency < 25% — knowledge isn't sticking.
		if w.AvgConsistency < 0.25 && w.StudentsStruggling > 0 {
			recs = append(recs, rec(rc, 1, "schedule_review", w.Topic,
				fmt.Sprintf("Стабильность знаний %.0f%% по теме — материал не закрепляется без регулярного повторения",
					w.AvgConsistency*100),
				w.StudentsStruggling, "consistency_critical", w))
			continue
		}

		// Rule review_needed: 25-49% struggling — lighter than full quiz.
		if w.StudentsStruggling < half && w.StudentsStruggling > 0 {
			recs = append(recs, rec(rc, 2, "schedule_review", w.Topic,
				fmt.Sprintf("%d ученик(ов) с трудом справляется — назначьте повторение через 3 дня",
					w.StudentsStruggling),
				w.StudentsStruggling, "review_needed", w))
		}
	}

	// Rule at_risk: students inactive 3+ days.
	if atRisk > 0 {
		recs = append(recs, rec(rc, 1, "check_students", "", // no topic
			fmt.Sprintf("%d учеников не заходили 3+ дня — стоит написать им", atRisk),
			atRisk, "at_risk", nil))
	}

	// Rule tar_low: TAR below threshold with enough data.
	if rc.TARTotal >= 5 && rc.TAR < 0.65 {
		recs = append(recs, rec(rc, 2, "rate_questions", "",
			fmt.Sprintf("TAR %.0f%% — оцените сгенерированные вопросы, чтобы улучшить качество AI",
				rc.TAR*100),
			0, "tar_low", nil))
	}

	// Rule celebrate: class performing well.
	if ins.QuizStats.AvgScore >= 85 && atRisk == 0 && ins.StudentCount > 0 {
		recs = append(recs, rec(rc, 3, "celebrate", "",
			fmt.Sprintf("Средний балл %.0f%% — группа молодцы! Можно усложнить материал",
				ins.QuizStats.AvgScore),
			0, "celebrate", nil))
	}

	return recs
}

// rec is a constructor for AIRecommendation that snapshots rule_data as JSON.
func rec(
	rc RuleContext,
	priority int,
	action, topic, reason string,
	studentCount int,
	ruleKey string,
	snapshot interface{},
) *domain.AIRecommendation {
	data := []byte("{}")
	if snapshot != nil {
		if b, err := json.Marshal(snapshot); err == nil {
			data = b
		}
	}
	groupID := rc.GroupID
	return &domain.AIRecommendation{
		ID:           uuid.New(),
		TeacherID:    rc.TeacherID,
		GroupID:      &groupID,
		Priority:     priority,
		Action:       action,
		Topic:        topic,
		Reason:       reason,
		StudentCount: studentCount,
		RuleKey:      ruleKey,
		RuleData:     data,
		Status:       "pending",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
	}
}
