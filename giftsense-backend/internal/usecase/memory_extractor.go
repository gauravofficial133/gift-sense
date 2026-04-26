package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type memoriesResponse struct {
	Memories []domain.MemoryEvidence `json:"memories"`
}

func ExtractMemories(ctx context.Context, llm port.LLMClient, chunks []domain.Chunk, recipient domain.RecipientDetails) ([]domain.MemoryEvidence, error) {
	prompt := buildMemoryExtractionPrompt(chunks, recipient)
	raw, err := llm.Complete(ctx, prompt, port.CompletionOptions{
		JSONMode:     true,
		MaxTokens:    500,
		SystemPrompt: "You are a memory curator. Extract meaningful quotes and moments from conversations. Respond ONLY with valid JSON.",
	})
	if err != nil {
		return nil, fmt.Errorf("memory extraction: %w", err)
	}
	var resp memoriesResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, fmt.Errorf("memory extraction parse: %w", err)
	}
	if len(resp.Memories) == 0 {
		return nil, fmt.Errorf("memory extraction: no memories found")
	}
	return limitMemories(resp.Memories, 5), nil
}

func buildMemoryExtractionPrompt(chunks []domain.Chunk, recipient domain.RecipientDetails) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Extract 3-5 memorable quotes or moments from this conversation about/with %s", recipient.Name)
	if recipient.Relation != "" {
		fmt.Fprintf(&sb, " (%s)", recipient.Relation)
	}
	sb.WriteString(".\n\nConversation excerpts:\n")
	for _, c := range chunks {
		sb.WriteString("---\n")
		sb.WriteString(c.AnonymizedText)
		sb.WriteString("\n")
	}
	sb.WriteString(`
For each memory, provide:
- quote: the actual text or a close paraphrase (max 100 chars)
- context: brief context about when/why this was said (max 60 chars)
- emotion: the dominant emotion (e.g., "joy", "love", "humor", "nostalgia")

Respond with JSON: {"memories":[{"quote":"...","context":"...","emotion":"..."}]}`)
	return sb.String()
}

func limitMemories(memories []domain.MemoryEvidence, max int) []domain.MemoryEvidence {
	if len(memories) <= max {
		return memories
	}
	return memories[:max]
}
