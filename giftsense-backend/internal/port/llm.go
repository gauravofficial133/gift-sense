package port

import "context"

type CompletionOptions struct {
	MaxTokens    int
	JSONMode     bool
	SystemPrompt string // non-empty overrides the adapter's default system prompt
}

type LLMClient interface {
	Complete(ctx context.Context, prompt string, opts CompletionOptions) (string, error)
}
