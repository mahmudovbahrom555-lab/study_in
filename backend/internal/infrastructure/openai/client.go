// Package openai wraps the go-openai SDK with domain-friendly types.
package openai

import (
	"context"
	"fmt"

	goopenai "github.com/sashabaranov/go-openai"
)

// Client is a thin wrapper over go-openai with cost tracking helpers.
type Client struct {
	c     *goopenai.Client
	model string
	embed string
}

// NewClient creates a new Client. model and embed default to gpt-4o-mini /
// text-embedding-3-small when empty.
func NewClient(apiKey, model, embedModel string) *Client {
	if model == "" {
		model = ModelGPT4oMini
	}
	if embedModel == "" {
		embedModel = ModelEmbedSmall
	}
	return &Client{c: goopenai.NewClient(apiKey), model: model, embed: embedModel}
}

// IsConfigured returns false when the API key is not set (graceful degradation).
func (c *Client) IsConfigured() bool {
	return c != nil && c.c != nil
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

type ChatMessage struct {
	Role    string
	Content string
}

type ChatRequest struct {
	SystemPrompt string
	Messages     []ChatMessage
	// Override default model (optional).
	Model       string
	MaxTokens   int
	Temperature float32
	// JSONMode forces the response to be valid JSON (response_format: json_object).
	JSONMode bool
}

type ChatResponse struct {
	Content   string
	TokensIn  int
	TokensOut int
	Model     string
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = c.model
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2000
	}

	var msgs []goopenai.ChatCompletionMessage
	if req.SystemPrompt != "" {
		msgs = append(msgs, goopenai.ChatCompletionMessage{
			Role:    goopenai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, goopenai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	creq := goopenai.ChatCompletionRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	}
	if req.JSONMode {
		creq.ResponseFormat = &goopenai.ChatCompletionResponseFormat{
			Type: goopenai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	resp, err := c.c.CreateChatCompletion(ctx, creq)
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: no choices returned")
	}

	return &ChatResponse{
		Content:   resp.Choices[0].Message.Content,
		TokensIn:  resp.Usage.PromptTokens,
		TokensOut: resp.Usage.CompletionTokens,
		Model:     resp.Model,
	}, nil
}

// ─── Embeddings ───────────────────────────────────────────────────────────────

type EmbedResponse struct {
	Vector   []float32
	TokensIn int
}

func (c *Client) Embed(ctx context.Context, text string) (*EmbedResponse, error) {
	resp, err := c.c.CreateEmbeddings(ctx, goopenai.EmbeddingRequest{
		Input: []string{text},
		Model: goopenai.EmbeddingModel(c.embed),
	})
	if err != nil {
		return nil, fmt.Errorf("openai embed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("openai: no embedding data")
	}
	return &EmbedResponse{
		Vector:   resp.Data[0].Embedding,
		TokensIn: resp.Usage.PromptTokens,
	}, nil
}

// EmbedBatch embeds multiple texts in a single API call.
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, int, error) {
	resp, err := c.c.CreateEmbeddings(ctx, goopenai.EmbeddingRequest{
		Input: texts,
		Model: goopenai.EmbeddingModel(c.embed),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("openai embed batch: %w", err)
	}
	result := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		result[i] = d.Embedding
	}
	return result, resp.Usage.PromptTokens, nil
}
