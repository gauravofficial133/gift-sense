package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/port"
)

type AssetPlanRequest struct {
	Style   string   `json:"style"`
	Subject string   `json:"subject"`
	Colors  []string `json:"colors"`
	Purpose string   `json:"purpose"`
}

type AssetPlanResult struct {
	RefinedPrompt string `json:"refined_prompt"`
}

type AssetPromptPlanner struct {
	llm port.LLMClient
}

func NewAssetPromptPlanner(llm port.LLMClient) *AssetPromptPlanner {
	return &AssetPromptPlanner{llm: llm}
}

func (p *AssetPromptPlanner) RefinePrompt(ctx context.Context, req AssetPlanRequest) (*AssetPlanResult, error) {
	var sb strings.Builder
	sb.WriteString("You are an expert at writing DALL-E 3 image generation prompts for greeting card assets.\n\n")
	sb.WriteString("Given the following structured inputs, write a single optimized DALL-E prompt.\n")
	sb.WriteString("The result should be a transparent-background decorative illustration suitable for a greeting card.\n\n")
	fmt.Fprintf(&sb, "Style: %s\n", req.Style)
	fmt.Fprintf(&sb, "Subject: %s\n", req.Subject)
	if len(req.Colors) > 0 {
		fmt.Fprintf(&sb, "Color palette: %s\n", strings.Join(req.Colors, ", "))
	}
	fmt.Fprintf(&sb, "Purpose: %s\n", req.Purpose)
	sb.WriteString("\nRespond with ONLY the refined prompt text, no explanation, no quotes.")

	raw, err := p.llm.Complete(ctx, sb.String(), port.CompletionOptions{
		MaxTokens:    200,
		SystemPrompt: "You write concise, effective DALL-E 3 prompts. Respond with only the prompt text.",
	})
	if err != nil {
		return nil, fmt.Errorf("refine prompt: %w", err)
	}

	return &AssetPlanResult{RefinedPrompt: strings.TrimSpace(raw)}, nil
}
