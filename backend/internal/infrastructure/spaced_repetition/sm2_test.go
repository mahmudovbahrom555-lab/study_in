package spaced_repetition_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	sm2 "github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/spaced_repetition"
)

func freshMastery() *domain.TopicMastery {
	return &domain.TopicMastery{
		ID:               uuid.New(),
		StudentID:        uuid.New(),
		TopicID:          uuid.New(),
		IntervalDays:     1,
		EaseFactor:       2.5,
		ConfidenceScore:  0.5,
		ConsistencyScore: 0.5,
		CorrectStreak:    0,
		NextReview:       time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func TestUpdate_WrongAnswer_ResetsInterval(t *testing.T) {
	m := freshMastery()
	m.IntervalDays = 15
	sm2.Update(m, false, 0)
	if m.IntervalDays != 1 {
		t.Fatalf("expected interval=1 after wrong answer, got %d", m.IntervalDays)
	}
}

func TestUpdate_CorrectAnswer_AdvancesInterval(t *testing.T) {
	m := freshMastery()
	sm2.Update(m, true, sm2.QualityGood)
	if m.IntervalDays <= 1 {
		t.Fatalf("expected interval>1 after correct, got %d", m.IntervalDays)
	}
}

func TestUpdate_ConfidenceRisesOnCorrect(t *testing.T) {
	m := freshMastery()
	before := m.ConfidenceScore
	sm2.Update(m, true, sm2.QualityEasy)
	if m.ConfidenceScore <= before {
		t.Fatalf("confidence should rise on correct: %.3f → %.3f", before, m.ConfidenceScore)
	}
}

func TestUpdate_ConfidenceFallsOnWrong(t *testing.T) {
	m := freshMastery()
	m.ConfidenceScore = 0.9
	sm2.Update(m, false, 0)
	if m.ConfidenceScore >= 0.9 {
		t.Fatalf("confidence should fall on wrong: %.3f", m.ConfidenceScore)
	}
}

func TestUpdate_ConsistencyPeaksAfterStreak(t *testing.T) {
	m := freshMastery()
	for i := 0; i < 5; i++ {
		sm2.Update(m, true, sm2.QualityGood)
	}
	if m.ConsistencyScore < 0.99 {
		t.Fatalf("consistency should be ~1.0 after 5 correct, got %.3f", m.ConsistencyScore)
	}
}

func TestUpdate_ConsistencyHalvedOnWrongAfterStreak(t *testing.T) {
	m := freshMastery()
	for i := 0; i < 5; i++ {
		sm2.Update(m, true, sm2.QualityGood)
	}
	before := m.ConsistencyScore
	sm2.Update(m, false, 0)
	if m.ConsistencyScore >= before {
		t.Fatalf("consistency should drop on wrong after streak: %.3f → %.3f", before, m.ConsistencyScore)
	}
	if m.ConsistencyScore < before*0.4 || m.ConsistencyScore > before*0.6 {
		t.Fatalf("consistency should be ~50%% of previous: %.3f → %.3f", before, m.ConsistencyScore)
	}
}

func TestUpdate_CorrectStreakResetsOnWrong(t *testing.T) {
	m := freshMastery()
	sm2.Update(m, true, sm2.QualityGood)
	sm2.Update(m, true, sm2.QualityGood)
	if m.CorrectStreak != 2 {
		t.Fatalf("expected streak=2, got %d", m.CorrectStreak)
	}
	sm2.Update(m, false, 0)
	if m.CorrectStreak != 0 {
		t.Fatalf("expected streak=0 after wrong, got %d", m.CorrectStreak)
	}
}

func TestIsDue_NotDueAfterUpdate(t *testing.T) {
	m := freshMastery()
	sm2.Update(m, true, sm2.QualityGood)
	if sm2.IsDue(m) {
		t.Fatal("should not be due immediately after advancing interval")
	}
}
