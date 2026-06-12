package prompts

import (
	"encoding/json"
	"fmt"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const RecommendationExplainerSystem = `You are a concise learning analytics advisor for English language teachers.
Given a specific teaching recommendation and its supporting metrics, explain in 3-4 sentences:
1. Why this recommendation was triggered (reference the specific numbers).
2. What outcome the teacher can realistically expect if they act on it.
3. One concrete first step.
Write in Russian. No bullet points — flowing prose. Under 80 words.`

// RecommendationExplainerPrompt builds the user prompt for GPT to explain a recommendation.
func RecommendationExplainerPrompt(rec *domain.AIRecommendation) string {
	dataStr := string(rec.RuleData)
	if dataStr == "" || dataStr == "{}" {
		dataStr = "нет дополнительных данных"
	}

	// Pretty-print the rule data if it's valid JSON.
	var parsed interface{}
	if json.Unmarshal(rec.RuleData, &parsed) == nil {
		if b, err := json.MarshalIndent(parsed, "", "  "); err == nil {
			dataStr = string(b)
		}
	}

	return fmt.Sprintf(`Рекомендация для учителя:
Действие: %s
Тема: %s
Причина (краткая): %s
Метрики на момент формирования рекомендации:
%s

Объясни учителю, почему эта рекомендация важна и что произойдёт, если он последует ей.`,
		rec.Action, rec.Topic, rec.Reason, dataStr)
}
