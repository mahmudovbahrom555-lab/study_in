package prompts

// QuizGeneratorSystem instructs GPT to return deterministic JSON only.
// The model must never deviate from the schema — JSONMode must be enabled.
const QuizGeneratorSystem = `You are an educational quiz generator.
Return ONLY valid JSON with no markdown, no explanations, no extra text.

JSON schema:
{
  "title": "Quiz title derived from content",
  "questions": [
    {
      "text": "Question text",
      "topic_tags": ["topic1", "topic2"],
      "difficulty": "A1|A2|B1|B2|C1|C2",
      "options": [
        {"text": "Option text", "is_correct": true},
        {"text": "Option text", "is_correct": false},
        {"text": "Option text", "is_correct": false},
        {"text": "Option text", "is_correct": false}
      ],
      "explanation": "One sentence explaining the correct answer"
    }
  ]
}

Rules:
- Exactly 4 options per question, exactly 1 correct.
- topic_tags: 2-4 specific topics from the material (2-3 words each).
- difficulty: CEFR level matching the linguistic/cognitive demand of the question.
- explanation: one concise sentence, in the same language as the material.
- Generate exactly the requested number of questions.`

// QuizGeneratorUserPrompt builds the user message for quiz generation.
func QuizGeneratorUserPrompt(context string, numQuestions int, cefrLevel, subject string) string {
	level := cefrLevel
	if level == "" {
		level = "B1"
	}
	subj := subject
	if subj == "" {
		subj = "general"
	}
	return "Generate " + itoa(numQuestions) + " multiple-choice questions for " + subj +
		" at CEFR level " + level + " based on the following material:\n\n---\n" + context + "\n---"
}

func itoa(n int) string {
	if n <= 0 {
		return "20"
	}
	// simple int-to-string without importing strconv at package level
	b := make([]byte, 0, 3)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
