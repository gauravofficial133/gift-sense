package anthropic

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/giftsense/backend/internal/port"
)

type LLMClient struct {
	client    anthropic.Client
	model     anthropic.Model
	maxTokens int
}

func NewLLMClient(apiKey, model string, maxTokens int) (*LLMClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is required")
	}
	if model == "" {
		return nil, fmt.Errorf("Anthropic model is required")
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &LLMClient{client: client, model: anthropic.Model(model), maxTokens: maxTokens}, nil
}

func (c *LLMClient) Complete(ctx context.Context, prompt string, opts port.CompletionOptions) (string, error) {
	maxTokens := c.maxTokens
	if opts.MaxTokens > 0 {
		maxTokens = opts.MaxTokens
	}

	systemPrompt := opts.SystemPrompt
	if opts.JSONMode && systemPrompt == "" {
		systemPrompt = "Respond ONLY with valid JSON. No markdown fences, no explanation, no extra text."
	} else if opts.JSONMode {
		systemPrompt += "\nRespond ONLY with valid JSON. No markdown fences, no explanation, no extra text."
	}

	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: int64(maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	resp, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("anthropic message creation failed: %w", err)
	}
	if len(resp.Content) == 0 {
		return "", fmt.Errorf("anthropic returned no content blocks")
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			text := block.Text
			if opts.JSONMode {
				text = stripMarkdownFences(text)
			}
			return text, nil
		}
	}
	return "", fmt.Errorf("anthropic returned no text content")
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	return s
}
