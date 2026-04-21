package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type CardGenerator struct {
	llm port.LLMClient
}

func NewCardGenerator(llm port.LLMClient) *CardGenerator {
	return &CardGenerator{llm: llm}
}

type cardLLMResponse struct {
	Body      string `json:"body"`
	Closing   string `json:"closing"`
	Signature string `json:"signature"`
}

func (g *CardGenerator) Generate(ctx context.Context, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal) (domain.CardRender, error) {
	occKey := DetectOccasion(recipient.Occasion)
	tmpl := GetOccasionTemplate(occKey)
	group := DetectEmotionGroup(emotions)
	palette := GetEmotionPalette(group)
	headline := fmt.Sprintf(tmpl.Greeting, recipient.Name)

	insightParts := make([]string, 0, 3)
	for i, ins := range insights {
		if i >= 3 {
			break
		}
		insightParts = append(insightParts, ins.Insight)
	}
	insightsSummary := strings.Join(insightParts, "; ")

	emotionParts := make([]string, 0, len(emotions))
	for _, e := range emotions {
		emotionParts = append(emotionParts, fmt.Sprintf("%s (%.0f%%)", e.Name, e.Intensity*100))
	}
	emotionsSummary := strings.Join(emotionParts, ", ")

	prompt := fmt.Sprintf(`You write the personal message inside a greeting card.

Occasion: %s
Recipient: %s
Detected emotional tone: %s
Personality insights: %s

Write a warm, personal card message that matches the emotional tone above.
If the tone is playful or humorous, be light and funny.
If the tone is tender or loving, be heartfelt and sincere.
If the tone is nostalgic or warm, be reflective and cozy.

Rules:
- Body: 2-3 sentences, 240 characters max. No quotes from chats. Grounded in the insights.
- Closing: 5 words max (e.g., "With all my heart", "Always yours").
- Signature: short line like "With love," or "Yours always,".
- ASCII and common punctuation only. No emoji.

Respond in JSON: {"body":"...","closing":"...","signature":"..."}`,
		recipient.Occasion, recipient.Name, emotionsSummary, insightsSummary)

	raw, err := g.llm.Complete(ctx, prompt, port.CompletionOptions{JSONMode: true})
	if err != nil {
		return domain.CardRender{}, fmt.Errorf("card generator: llm complete: %w", err)
	}

	var resp cardLLMResponse
	if parseErr := json.Unmarshal([]byte(raw), &resp); parseErr != nil {
		return domain.CardRender{}, fmt.Errorf("card generator: parse response: %w", parseErr)
	}

	content := domain.CardContent{
		Headline:    headline,
		Body:        resp.Body,
		Closing:     resp.Closing,
		Signature:   resp.Signature,
		Recipient:   recipient.Name,
		Emotions:    emotions,
		OccasionKey: string(occKey),
	}

	svgStr, err := RenderSVG(tmpl.Motif, palette, content)
	if err != nil {
		return domain.CardRender{}, fmt.Errorf("card generator: %w", err)
	}

	pdfBytes, err := RenderPDF(palette, content)
	if err != nil {
		return domain.CardRender{}, fmt.Errorf("card generator: %w", err)
	}

	return domain.CardRender{
		SVG:       svgStr,
		PDFBase64: base64.StdEncoding.EncodeToString(pdfBytes),
		ThemeID:   string(occKey),
		Content:   content,
	}, nil
}
