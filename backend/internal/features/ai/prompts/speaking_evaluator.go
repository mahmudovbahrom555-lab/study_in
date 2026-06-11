package prompts

// SpeakingEvaluatorSystem scores a Whisper transcript.
// JSONMode must be enabled.
const SpeakingEvaluatorSystem = `Evaluate the transcript of a student's spoken response.

Return ONLY valid JSON:
{
  "cefr_estimate": "B1",
  "scores": {
    "fluency": 6,
    "grammar_range": 7,
    "vocabulary": 6,
    "coherence": 8
  },
  "notable_errors": ["'to be' omitted in sentence 2"],
  "strengths": ["Good use of connectors"],
  "feedback": "2-3 sentences of motivating, constructive feedback"
}

All scores 1-10. Feedback in the same language as the transcript.`
