package openai

// Model identifiers used across the AI layer.
const (
	ModelGPT4oMini      = "gpt-4o-mini"
	ModelGPT4o          = "gpt-4o"
	ModelEmbedSmall     = "text-embedding-3-small"
	ModelWhisper1       = "whisper-1"
	ModelTTS1           = "tts-1"
	ModelDallE3         = "dall-e-3"
)

// Cost per 1 million tokens in USD (as of mid-2025).
var costPer1MTokens = map[string][2]float64{
	// [input, output]
	ModelGPT4oMini:  {0.15, 0.60},
	ModelGPT4o:      {5.00, 15.00},
	ModelEmbedSmall: {0.02, 0.00},
}

// CalcCostUSD returns the cost in USD for the given token counts.
func CalcCostUSD(model string, tokensIn, tokensOut int) float64 {
	prices, ok := costPer1MTokens[model]
	if !ok {
		return 0
	}
	return float64(tokensIn)/1_000_000*prices[0] + float64(tokensOut)/1_000_000*prices[1]
}
