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

type TemplateSelection struct {
	TemplateID           string            `json:"template_id"`
	PaletteName          string            `json:"palette"`
	FontChoices          map[string]string `json:"font_choices,omitempty"`
	BackgroundChoice     int               `json:"background_choice"`
	GenerateIllustration bool              `json:"generate_illustration"`
	IllustrationSlot     string            `json:"illustration_slot,omitempty"`
}

type CardCopy struct {
	Headline           string `json:"headline"`
	Body               string `json:"body"`
	Closing            string `json:"closing"`
	Signature          string `json:"signature"`
	IllustrationPrompt string `json:"illustration_prompt,omitempty"`
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
	writeTemplateList(&sb, templates)
	writePaletteList(&sb)
	writeRecipientContext(&sb, recipient, insights, emotions, variation)
	sb.WriteString(`
RULES:
- Pick a template_id from the templates listed above.
- Pick a palette from the allowed palettes for that template, or any palette if the template uses "ai_decides" mode.
- If the template has font options per zone, pick one font per zone and include in font_choices (zone_id -> font_name).
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

func SelectTemplate(ctx context.Context, llm port.LLMClient, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string, templates []domain.TemplateDefinition) (TemplateSelection, error) {
	prompt := buildTemplateSelectionPrompt(recipient, insights, emotions, variation, templates)
	raw, err := llm.Complete(ctx, prompt, port.CompletionOptions{
		JSONMode:     true,
		MaxTokens:    300,
		SystemPrompt: "You are a visual designer for premium greeting cards. Choose the best template and visual settings. Respond ONLY with valid JSON.",
	})
	if err != nil {
		return TemplateSelection{}, fmt.Errorf("template selection: %w", err)
	}
	var sel TemplateSelection
	if err := json.Unmarshal([]byte(raw), &sel); err != nil {
		return TemplateSelection{}, fmt.Errorf("template selection parse: %w", err)
	}
	if sel.TemplateID == "" {
		return TemplateSelection{}, fmt.Errorf("template selection: empty template_id")
	}
	return sel, nil
}

func buildTemplateSelectionPrompt(recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string, templates []domain.TemplateDefinition) string {
	var sb strings.Builder
	sb.WriteString("You are a visual designer. Given the recipient context, choose the best template and visual settings. Do NOT write any card text.\n\n")
	writeTemplateList(&sb, templates)
	writePaletteList(&sb)
	writeRecipientContext(&sb, recipient, insights, emotions, variation)
	sb.WriteString(`
RULES:
- Pick a template_id from the templates above that BEST FITS the occasion and emotions.
- Prefer templates whose element layout matches the need: pick templates with illustration slots when a visual illustration would enhance the card; pick templates with data slots if data fields are relevant.
- Pick a palette that matches the occasion and emotions from the allowed palettes (or any if ai_decides mode).
- For each text zone with font options, include font_choices mapping zone_id -> font_name.
- If the template has multiple background options, pick via background_choice (0-indexed).
- Set generate_illustration to true ONLY if the chosen template has an illustration slot.
- If generating, set illustration_slot to match the slot_name from the template (e.g. "hero").

Respond ONLY with JSON:
{"template_id":"...","palette":"...","font_choices":{},"background_choice":0,"generate_illustration":false,"illustration_slot":"hero"}`)
	return sb.String()
}

func WriteCopy(ctx context.Context, llm port.LLMClient, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string, sel TemplateSelection, tplDef domain.TemplateDefinition) (CardCopy, error) {
	prompt := buildCopywritingPrompt(recipient, insights, emotions, variation, sel, tplDef)
	raw, err := llm.Complete(ctx, prompt, port.CompletionOptions{
		JSONMode:     true,
		MaxTokens:    400,
		SystemPrompt: "You are a premium greeting card copywriter. Write warm, personal card text. Respond ONLY with valid JSON.",
	})
	if err != nil {
		return CardCopy{}, fmt.Errorf("copywriting: %w", err)
	}
	var copy CardCopy
	if err := json.Unmarshal([]byte(raw), &copy); err != nil {
		return CardCopy{}, fmt.Errorf("copywriting parse: %w", err)
	}
	if copy.Body == "" {
		return CardCopy{}, fmt.Errorf("copywriting: empty body")
	}
	trimCopySafe(&copy)
	return copy, nil
}

func buildCopywritingPrompt(recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string, sel TemplateSelection, tplDef domain.TemplateDefinition) string {
	var sb strings.Builder
	sb.WriteString("Write greeting card text for the template and context below.\n\n")
	fmt.Fprintf(&sb, "TEMPLATE: %s (%s)\n", tplDef.ID, tplDef.Name)
	fmt.Fprintf(&sb, "PALETTE: %s\n", sel.PaletteName)
	writeTextZoneConstraints(&sb, tplDef)
	writeRecipientContext(&sb, recipient, insights, emotions, variation)
	sb.WriteString(`
RULES:
- Write text that fits WITHIN the char limits specified for each text zone above.
- Headline: greeting for the occasion with recipient name. Respect the headline zone's char_max.
- Body: personal message grounded in the personality insights. Respect the body zone's char_max. Use the tone specified for that zone.
- Closing: short warm closing. Respect the closing zone's char_max.
- Signature: short line. Respect the signature zone's char_max.
- ASCII and common punctuation only. No emoji.
- If the template has an illustration slot, provide illustration_prompt: a DALL-E prompt matching the slot's style_hint (transparent background, no text). Max 200 chars.
- Write text that looks good at the font sizes specified for each zone.

Respond ONLY with JSON:
{"headline":"...","body":"...","closing":"...","signature":"...","illustration_prompt":""}`)
	return sb.String()
}

func writeTemplateList(sb *strings.Builder, templates []domain.TemplateDefinition) {
	if len(templates) == 0 {
		return
	}
	sb.WriteString("AVAILABLE TEMPLATES:\n")
	for _, tpl := range templates {
		fmt.Fprintf(sb, "- %s: %s.", tpl.ID, tpl.Name)
		if tpl.Family != "" {
			fmt.Fprintf(sb, " Family: %s.", tpl.Family)
		}
		if tpl.Canvas.Orientation != "" {
			fmt.Fprintf(sb, " %s %dx%d.", strings.ToUpper(tpl.Canvas.Orientation[:1])+tpl.Canvas.Orientation[1:], tpl.Canvas.Width, tpl.Canvas.Height)
		}
		writeTemplateMeta(sb, tpl)
		writeElementSummary(sb, tpl)
		sb.WriteString("\n")
	}
}

func writeTemplateMeta(sb *strings.Builder, tpl domain.TemplateDefinition) {
	if len(tpl.Occasions) > 0 {
		strs := make([]string, len(tpl.Occasions))
		for i, o := range tpl.Occasions {
			strs[i] = string(o)
		}
		fmt.Fprintf(sb, " Occasions: %s.", strings.Join(strs, ", "))
	}
	if len(tpl.Emotions) > 0 {
		strs := make([]string, len(tpl.Emotions))
		for i, e := range tpl.Emotions {
			strs[i] = string(e)
		}
		fmt.Fprintf(sb, " Emotions: %s.", strings.Join(strs, ", "))
	}
	if tpl.VariationRules.PaletteMode == "set" && len(tpl.VariationRules.AllowedPalettes) > 0 {
		fmt.Fprintf(sb, " Palettes: %s.", strings.Join(tpl.VariationRules.AllowedPalettes, ", "))
	}
	if len(tpl.VariationRules.BackgroundOptions) > 1 {
		fmt.Fprintf(sb, " %d background options.", len(tpl.VariationRules.BackgroundOptions))
	}
}

func writeElementSummary(sb *strings.Builder, tpl domain.TemplateDefinition) {
	var textZones, photos, illustrations, dataSlots, decoratives int
	for _, el := range tpl.Elements {
		switch el.Type {
		case "text_zone":
			textZones++
		case "photo_slot":
			photos++
		case "ai_illustration_slot", "illustration_slot":
			illustrations++
		case "data_slot":
			dataSlots++
		case "decorative":
			decoratives++
		}
	}
	parts := make([]string, 0, 5)
	if textZones > 0 {
		parts = append(parts, fmt.Sprintf("%d text zones", textZones))
	}
	if photos > 0 {
		parts = append(parts, fmt.Sprintf("%d photo slots", photos))
	}
	if illustrations > 0 {
		parts = append(parts, fmt.Sprintf("%d illustration slots", illustrations))
	}
	if dataSlots > 0 {
		parts = append(parts, fmt.Sprintf("%d data slots", dataSlots))
	}
	if decoratives > 0 {
		parts = append(parts, fmt.Sprintf("%d decoratives", decoratives))
	}
	if len(parts) > 0 {
		fmt.Fprintf(sb, " Layout: %s.", strings.Join(parts, ", "))
	}

	for _, el := range tpl.Elements {
		if el.TextZone != nil {
			fmt.Fprintf(sb, " [%s: role=%s, %d-%d chars", el.ID, el.TextZone.SemanticRole, el.TextZone.CharMin, el.TextZone.CharMax)
			if el.TextZone.Purpose != "" {
				fmt.Fprintf(sb, ", purpose=%q", el.TextZone.Purpose)
			}
			if len(el.TextZone.FontOptions) > 0 {
				fmt.Fprintf(sb, ", fonts=%s", strings.Join(el.TextZone.FontOptions, "/"))
			}
			sb.WriteString("]")
		}
		if el.IllustrationSlot != nil {
			if el.Size != nil {
				fmt.Fprintf(sb, " [%s: illustration, slot=%s, style=%s, %dx%d]",
					el.ID, el.IllustrationSlot.SlotName, el.IllustrationSlot.StyleHint,
					el.Size.W, el.Size.H)
			} else {
				fmt.Fprintf(sb, " [%s: illustration, slot=%s, style=%s]",
					el.ID, el.IllustrationSlot.SlotName, el.IllustrationSlot.StyleHint)
			}
		}
		if el.DataSlot != nil {
			fmt.Fprintf(sb, " [%s: data, field=%s]", el.ID, el.DataSlot.Field)
		}
		if el.PhotoSlot != nil {
			fmt.Fprintf(sb, " [%s: photo, shape=%s]", el.ID, el.PhotoSlot.Shape)
		}
	}
}

func writePaletteList(sb *strings.Builder) {
	sb.WriteString("\nAVAILABLE PALETTES:\n")
	for name, desc := range paletteDescriptions {
		fmt.Fprintf(sb, "- %s: %s\n", name, desc)
	}
}

func writeRecipientContext(sb *strings.Builder, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, variation string) {
	fmt.Fprintf(sb, "\nRECIPIENT: %s", recipient.Name)
	if recipient.Relation != "" {
		fmt.Fprintf(sb, " (%s)", recipient.Relation)
	}
	fmt.Fprintf(sb, "\nOCCASION: %s\n", recipient.Occasion)
	if len(insights) > 0 {
		parts := make([]string, 0, 3)
		for i, ins := range insights {
			if i >= 3 {
				break
			}
			parts = append(parts, ins.Insight)
		}
		fmt.Fprintf(sb, "PERSONALITY: %s\n", strings.Join(parts, "; "))
	}
	if len(emotions) > 0 {
		parts := make([]string, 0, len(emotions))
		for _, e := range emotions {
			parts = append(parts, fmt.Sprintf("%s (%.0f%%)", e.Name, e.Intensity*100))
		}
		fmt.Fprintf(sb, "EMOTIONAL TONE: %s\n", strings.Join(parts, ", "))
	}
	fmt.Fprintf(sb, "VARIATION STYLE: %s\n", variation)
}

func writeTextZoneConstraints(sb *strings.Builder, tplDef domain.TemplateDefinition) {
	sb.WriteString("\nTEXT ZONE CONSTRAINTS:\n")
	for _, el := range tplDef.Elements {
		if el.TextZone == nil {
			continue
		}
		tz := el.TextZone
		fmt.Fprintf(sb, "- %s (%s): %d-%d chars, tone: %s, role: %s",
			el.ID, tz.Purpose, tz.CharMin, tz.CharMax, tz.Tone, tz.SemanticRole)
		if len(tz.FontOptions) > 0 {
			fmt.Fprintf(sb, ", fonts: %s", strings.Join(tz.FontOptions, "/"))
		}
		if el.Size != nil {
			fmt.Fprintf(sb, ", area: %dx%dpx", el.Size.W, el.Size.H)
		}
		fmt.Fprintf(sb, ", font-size: %d-%dpx", tz.FontSizeRange.Min, tz.FontSizeRange.Max)
		sb.WriteString("\n")
	}

	hasIllustration := false
	for _, el := range tplDef.Elements {
		if el.IllustrationSlot != nil {
			hasIllustration = true
			if el.Size != nil {
				fmt.Fprintf(sb, "- %s (illustration): slot=%s, style-hint=%s, %dx%dpx\n",
					el.ID, el.IllustrationSlot.SlotName, el.IllustrationSlot.StyleHint, el.Size.W, el.Size.H)
			} else {
				fmt.Fprintf(sb, "- %s (illustration): slot=%s, style-hint=%s\n",
					el.ID, el.IllustrationSlot.SlotName, el.IllustrationSlot.StyleHint)
			}
		}
	}
	if hasIllustration {
		sb.WriteString("  -> This template has an illustration slot. Provide illustration_prompt if appropriate.\n")
	}

	for _, el := range tplDef.Elements {
		if el.DataSlot != nil {
			fmt.Fprintf(sb, "- %s (data): field=%s, format=%s\n",
				el.ID, el.DataSlot.Field, el.DataSlot.FormatTemplate)
		}
	}
}

func trimCopySafe(copy *CardCopy) {
	if len(copy.Headline) > 40 {
		copy.Headline = copy.Headline[:40]
	}
	if len(copy.Body) > 240 {
		copy.Body = copy.Body[:240]
	}
	if len(copy.Closing) > 30 {
		copy.Closing = copy.Closing[:30]
	}
	if len(copy.Signature) > 30 {
		copy.Signature = copy.Signature[:30]
	}
}

func MergeArtDirection(sel TemplateSelection, copy CardCopy) ArtDirection {
	return ArtDirection{
		TemplateID:           sel.TemplateID,
		RecipeID:             sel.TemplateID,
		PaletteName:          sel.PaletteName,
		Headline:             copy.Headline,
		Body:                 copy.Body,
		Closing:              copy.Closing,
		Signature:            copy.Signature,
		GenerateIllustration: sel.GenerateIllustration,
		IllustrationPrompt:   copy.IllustrationPrompt,
		IllustrationSlot:     sel.IllustrationSlot,
		FontChoices:          sel.FontChoices,
		BackgroundChoice:     sel.BackgroundChoice,
	}
}
