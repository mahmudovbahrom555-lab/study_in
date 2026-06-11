// Package spaced_repetition implements the SM-2 algorithm used by Anki/Duolingo.
// After each quiz answer the caller updates TopicMastery via Update().
package spaced_repetition

import (
	"math"
	"time"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Quality values passed to Update (mirrors SM-2 scale 0-5).
const (
	QualityBlackout  = 0 // complete blackout
	QualityBad       = 1 // wrong but familiar on seeing answer
	QualityFail      = 2 // wrong but easy to remember
	QualityHard      = 3 // correct with difficulty
	QualityGood      = 4 // correct after hesitation
	QualityEasy      = 5 // perfect response
)

// Update applies one SM-2 iteration to mastery.
// correct=true  → increase interval (spaced repetition reward)
// correct=false → reset to 1 day (re-learn)
// quality is only used when correct=true to adjust ease factor.
func Update(mastery *domain.TopicMastery, correct bool, quality int) {
	if !correct {
		mastery.IntervalDays = 1
		mastery.EaseFactor = math.Max(1.3, mastery.EaseFactor-0.2)
	} else {
		switch mastery.IntervalDays {
		case 1:
			mastery.IntervalDays = 6
		case 6:
			mastery.IntervalDays = 15
		default:
			mastery.IntervalDays = int(math.Round(float64(mastery.IntervalDays) * mastery.EaseFactor))
		}
		// EF' = EF + (0.1 - (5-q)*(0.08 + (5-q)*0.02))
		q := float64(quality)
		delta := 0.1 - (5-q)*(0.08+(5-q)*0.02)
		mastery.EaseFactor = math.Max(1.3, mastery.EaseFactor+delta)
	}
	mastery.NextReview = time.Now().AddDate(0, 0, mastery.IntervalDays)
	mastery.UpdatedAt = time.Now()
}

// IsDue returns true when the mastery item is ready for review.
func IsDue(mastery *domain.TopicMastery) bool {
	return !time.Now().Before(mastery.NextReview)
}
