package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

func (a *Analyzer) ExtractChatEmotions(ctx context.Context, chunks []domain.Chunk, recipient domain.RecipientDetails) ([]domain.EmotionSignal, error) {
	excerpts := chunks
	if len(excerpts) > 5 {
		excerpts = excerpts[:5]
	}

	var sb strings.Builder
	vocab := strings.Join(EmotionVocabulary, ", ")
	fmt.Fprintf(&sb, "System context: You are an emotion analyst. Return ONLY valid JSON. Emotion names must be from this vocabulary: %s. intensity is 0.0-1.0. emoji is a single unicode emoji matching the emotion.\n\n", vocab)
	fmt.Fprintf(&sb, "These are excerpts from the sender's conversation with %s. Identify the dominant emotions the SENDER feels TOWARD the recipient. Return max 5 emotions from the vocabulary as JSON array: [{\"name\":\"...\",\"emoji\":\"...\",\"intensity\":0.0}]\n\n", recipient.Name)

	for _, c := range excerpts {
		sb.WriteString("---\n")
		sb.WriteString(c.AnonymizedText)
		sb.WriteString("\n")
	}

	raw, err := a.llm.Complete(ctx, sb.String(), port.CompletionOptions{JSONMode: true})
	if err != nil {
		return []domain.EmotionSignal{{Name: "Warmth", Emoji: "🤗", Intensity: 0.7}}, nil
	}

	var signals []domain.EmotionSignal
	if parseErr := json.Unmarshal([]byte(raw), &signals); parseErr != nil {
		return []domain.EmotionSignal{{Name: "Warmth", Emoji: "🤗", Intensity: 0.7}}, nil
	}

	valid := make([]domain.EmotionSignal, 0, len(signals))
	vocabSet := make(map[string]bool, len(EmotionVocabulary))
	for _, v := range EmotionVocabulary {
		vocabSet[v] = true
	}
	for _, s := range signals {
		if vocabSet[s.Name] {
			valid = append(valid, s)
		}
	}

	if len(valid) == 0 {
		return []domain.EmotionSignal{{Name: "Warmth", Emoji: "🤗", Intensity: 0.7}}, nil
	}
	return valid, nil
}
