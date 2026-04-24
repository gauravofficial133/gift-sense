package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type ArtDirection struct {
	TemplateID           string            `json:"template_id"`
	RecipeID             string            `json:"recipe"`
	PaletteName          string            `json:"palette"`
	HeadlineFont         string            `json:"headline_font"`
	BodyFont             string            `json:"body_font"`
	Headline             string            `json:"headline"`
	Body                 string            `json:"body"`
	Closing              string            `json:"closing"`
	Signature            string            `json:"signature"`
	GenerateIllustration bool              `json:"generate_illustration"`
	IllustrationPrompt   string            `json:"illustration_prompt,omitempty"`
	IllustrationSlot     string            `json:"illustration_slot,omitempty"`
	FontChoices          map[string]string `json:"font_choices,omitempty"`
	BackgroundChoice     int               `json:"background_choice,omitempty"`
}

var paletteDescriptions = map[string]string{
	"sunrise_warmth":     "Warm peach/coral on cream. Emotions: warm.",
	"soft_rose_gold":     "Blush pink/mauve on soft pink. Emotions: tender, passionate.",
	"ocean_calm":         "Teal/navy on pale blue. Emotions: reflective.",
	"electric_joy":       "Orange/yellow on warm white. Emotions: joyful.",
	"midnight_elegant":   "Gold on deep navy. Emotions: reflective, neutral.",
	"forest_peace":       "Green/sage on mint. Emotions: warm, reflective.",
	"lavender_dream":     "Purple/lilac on lavender. Emotions: tender.",
	"golden_celebration": "Amber/gold on warm cream. Emotions: joyful, warm.",
}

func buildArtDirectionPromptWithTemplates(recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string, templates []domain.TemplateDefinition) string {
	var sb strings.Builder

	sb.WriteString("You are an art director for premium greeting cards. Given the recipient context below, choose the best template and color palette, then write the card copy.\n\n")

	if len(templates) > 0 {
		sb.WriteString("AVAILABLE TEMPLATES:\n")
		for _, tpl := range templates {
			fmt.Fprintf(&sb, "- %s: %s.", tpl.ID, tpl.Name)
			if tpl.Canvas.Orientation != "" {
				fmt.Fprintf(&sb, " %s %dx%d.", strings.ToUpper(tpl.Canvas.Orientation[:1])+tpl.Canvas.Orientation[1:], tpl.Canvas.Width, tpl.Canvas.Height)
			}
			if len(tpl.Occasions) > 0 {
				occasionStrs := make([]string, len(tpl.Occasions))
				for i, o := range tpl.Occasions {
					occasionStrs[i] = string(o)
				}
				fmt.Fprintf(&sb, " Occasions: %s.", strings.Join(occasionStrs, ", "))
			}
			if len(tpl.Emotions) > 0 {
				emotionStrs := make([]string, len(tpl.Emotions))
				for i, e := range tpl.Emotions {
					emotionStrs[i] = string(e)
				}
				fmt.Fprintf(&sb, " Emotions: %s.", strings.Join(emotionStrs, ", "))
			}
			if tpl.VariationRules.PaletteMode == "set" && len(tpl.VariationRules.AllowedPalettes) > 0 {
				fmt.Fprintf(&sb, " Palettes: %s.", strings.Join(tpl.VariationRules.AllowedPalettes, ", "))
			}
			for _, el := range tpl.Elements {
				if el.TextZone != nil && len(el.TextZone.FontOptions) > 0 {
					fmt.Fprintf(&sb, " Zone %s fonts: [%s].", el.ID, strings.Join(el.TextZone.FontOptions, ", "))
				}
			}
			if len(tpl.VariationRules.BackgroundOptions) > 1 {
				fmt.Fprintf(&sb, " %d background options.", len(tpl.VariationRules.BackgroundOptions))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\nAVAILABLE PALETTES:\n")
	for name, desc := range paletteDescriptions {
		fmt.Fprintf(&sb, "- %s: %s\n", name, desc)
	}

	fmt.Fprintf(&sb, "\nRECIPIENT: %s", recipient.Name)
	if recipient.Relation != "" {
		fmt.Fprintf(&sb, " (%s)", recipient.Relation)
	}
	fmt.Fprintf(&sb, "\nOCCASION: %s\n", recipient.Occasion)

	if len(insights) > 0 {
		parts := make([]string, 0, 3)
		for i, ins := range insights {
			if i >= 3 {
				break
			}
			parts = append(parts, ins.Insight)
		}
		fmt.Fprintf(&sb, "PERSONALITY: %s\n", strings.Join(parts, "; "))
	}

	if len(emotions) > 0 {
		parts := make([]string, 0, len(emotions))
		for _, e := range emotions {
			parts = append(parts, fmt.Sprintf("%s (%.0f%%)", e.Name, e.Intensity*100))
		}
		fmt.Fprintf(&sb, "EMOTIONAL TONE: %s\n", strings.Join(parts, ", "))
	}

	fmt.Fprintf(&sb, "VARIATION STYLE: %s\n", variation)

	sb.WriteString(`
RULES:
- Pick a template_id from the templates listed above.
- Pick a palette from the allowed palettes for that template, or any palette if the template uses "ai_decides" mode.
- If the template has font options per zone, pick one font per zone and include in font_choices (zone_id → font_name).
- If the template has multiple background options, pick one via background_choice (0-indexed).
- Headline: greeting for the occasion with recipient name. Max 40 chars.
- Body: 2-3 sentences, 240 chars max. Personal, grounded in the personality insights. No quotes from chats.
- Closing: 5 words max.
- Signature: short line.
- ASCII and common punctuation only. No emoji.
- If a custom illustration would enhance the card, set generate_illustration to true and provide:
  - illustration_prompt: a DALL-E prompt for a watercolor/flat-vector illustration (transparent background, no text). Max 200 chars.
  - illustration_slot: "hero", "corner", or "background". Default "hero".

Respond ONLY with JSON:
{"template_id":"...","palette":"...","font_choices":{},"background_choice":0,"headline":"...","body":"...","closing":"...","signature":"...","generate_illustration":false,"illustration_prompt":"","illustration_slot":"hero"}`)

	return sb.String()
}

func DirectArt(ctx context.Context, llm port.LLMClient, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string) (ArtDirection, error) {
	return DirectArtWithTemplates(ctx, llm, recipient, insights, emotions, variation, nil)
}

func DirectArtWithTemplates(ctx context.Context, llm port.LLMClient, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string, templates []domain.TemplateDefinition) (ArtDirection, error) {
	prompt := buildArtDirectionPromptWithTemplates(recipient, insights, emotions, variation, templates)

	raw, err := llm.Complete(ctx, prompt, port.CompletionOptions{
		JSONMode:     true,
		MaxTokens:    500,
		SystemPrompt: "You are a premium greeting card art director. Respond ONLY with valid JSON, no markdown fences, no extra text.",
	})
	if err != nil {
		return ArtDirection{}, fmt.Errorf("art direction: %w", err)
	}

	var dir ArtDirection
	if err := json.Unmarshal([]byte(raw), &dir); err != nil {
		return ArtDirection{}, fmt.Errorf("art direction parse: %w", err)
	}

	if dir.TemplateID != "" && dir.RecipeID == "" {
		dir.RecipeID = dir.TemplateID
	}
	if dir.RecipeID != "" && dir.TemplateID == "" {
		dir.TemplateID = dir.RecipeID
	}

	if dir.RecipeID == "" || dir.Body == "" {
		return ArtDirection{}, fmt.Errorf("art direction: incomplete response")
	}

	trimSafe(&dir)

	return dir, nil
}

func trimSafe(dir *ArtDirection) {
	if len(dir.Headline) > 40 {
		dir.Headline = dir.Headline[:40]
	}
	if len(dir.Body) > 240 {
		dir.Body = dir.Body[:240]
	}
	if len(dir.Closing) > 30 {
		dir.Closing = dir.Closing[:30]
	}
	if len(dir.Signature) > 30 {
		dir.Signature = dir.Signature[:30]
	}
}
