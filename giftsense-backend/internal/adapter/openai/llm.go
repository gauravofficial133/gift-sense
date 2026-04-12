package openai

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"

	"github.com/giftsense/backend/internal/port"
)

const systemPrompt = `You are a warm, insightful gift recommendation assistant for upahaar.ai who reads between the lines of conversations to understand people deeply.
RULES:
1. Only infer traits and suggest gifts supported by evidence in the provided conversation context.
2. Every gift suggestion MUST have an estimated price within the stated budget range.
3. Respond ONLY with a valid JSON object matching this schema: {"personality_insights": [{"insight": "string", "evidence_summary": "string"}], "gift_suggestions": [{"name": "string", "reason": "string", "estimated_price_inr": "string", "category": "string"}]}
4. Personality insights should be warm and human — written as a perceptive friend, not a corporate analyst.
5. Gift names must be specific enough to search for (e.g. "Pottery starter kit with air-dry clay" not "craft supplies").`

type LLMClient struct {
	client    openai.Client
	model     string
	maxTokens int
}

func NewLLMClient(apiKey, model string, maxTokens int) (*LLMClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("LLM model is required")
	}
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &LLMClient{client: client, model: model, maxTokens: maxTokens}, nil
}

func (llmClient *LLMClient) Complete(ctx context.Context, prompt string, opts port.CompletionOptions) (string, error) {
	maxTokens := llmClient.maxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	activeSystemPrompt := systemPrompt
	if opts.SystemPrompt != "" {
		activeSystemPrompt = opts.SystemPrompt
	}

	params := openai.ChatCompletionNewParams{
		Model:     shared.ChatModel(llmClient.model),
		MaxTokens: openai.Int(int64(maxTokens)),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(activeSystemPrompt),
			openai.UserMessage(prompt),
		},
	}

	if opts.JSONMode {
		params.ResponseFormat = openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &shared.ResponseFormatJSONObjectParam{},
		}
	}

	resp, err := llmClient.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("chat completion failed: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("chat completion returned no choices")
	}
	return resp.Choices[0].Message.Content, nil
}
