package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/giftsense/backend/internal/adapter/cardrender"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type CardGenerator struct {
	openaiLLM    port.LLMClient
	anthropicLLM port.LLMClient
	renderer     *cardrender.Renderer
	engine       *cardrender.TemplateEngine
	imageGen     port.ImageGenerator
	assetLib     *AssetLibrary
	tplStore     port.TemplateStore
	compiler     *HTMLCompiler
}

func NewCardGenerator(openaiLLM port.LLMClient, anthropicLLM port.LLMClient, renderer *cardrender.Renderer, engine *cardrender.TemplateEngine, imageGen port.ImageGenerator, assetLib *AssetLibrary) *CardGenerator {
	return &CardGenerator{openaiLLM: openaiLLM, anthropicLLM: anthropicLLM, renderer: renderer, engine: engine, imageGen: imageGen, assetLib: assetLib}
}

func (g *CardGenerator) SetTemplateStore(store port.TemplateStore, compiler *HTMLCompiler) {
	g.tplStore = store
	g.compiler = compiler
}

type cardJob struct {
	llm       port.LLMClient
	model     string
	variation string
}

func (g *CardGenerator) Generate(ctx context.Context, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal) []*domain.CardRender {
	if g.renderer == nil {
		return nil
	}

	jobs := g.buildJobs()
	if len(jobs) == 0 {
		return nil
	}

	var templates []domain.TemplateDefinition
	if g.tplStore != nil {
		if tpls, err := g.tplStore.List(ctx); err == nil && len(tpls) > 0 {
			templates = tpls
		}
	}

	if len(templates) == 0 {
		log.Println("no templates available, skipping card generation")
		return nil
	}

	var (
		mu    sync.Mutex
		cards []*domain.CardRender
		wg    sync.WaitGroup
	)

	for _, job := range jobs {
		wg.Add(1)
		go func(j cardJob) {
			defer wg.Done()
			card, err := g.generateOne(ctx, j, recipient, insights, emotions, templates)
			if err != nil {
				log.Printf("card generation (%s/%s): %v", j.model, j.variation, err)
				return
			}
			mu.Lock()
			cards = append(cards, card)
			mu.Unlock()
		}(job)
	}

	wg.Wait()

	occasionKey := string(DetectOccasion(recipient.Occasion))
	cards = ScoreCards(cards, occasionKey, emotions, templates)

	return cards
}

func (g *CardGenerator) buildJobs() []cardJob {
	var jobs []cardJob
	if g.openaiLLM != nil {
		jobs = append(jobs,
			cardJob{llm: g.openaiLLM, model: "openai", variation: "heartfelt and sincere"},
			cardJob{llm: g.openaiLLM, model: "openai", variation: "playful and creative"},
		)
	}
	if g.anthropicLLM != nil {
		jobs = append(jobs,
			cardJob{llm: g.anthropicLLM, model: "claude", variation: "elegant and poetic"},
			cardJob{llm: g.anthropicLLM, model: "claude", variation: "warm and conversational"},
		)
	}
	if len(jobs) == 0 && g.openaiLLM != nil {
		jobs = append(jobs, cardJob{llm: g.openaiLLM, model: "openai", variation: "heartfelt and sincere"})
	}
	return jobs
}

func (g *CardGenerator) generateOne(ctx context.Context, job cardJob, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, templates []domain.TemplateDefinition) (*domain.CardRender, error) {
	start := time.Now()

	sel, err := SelectTemplate(ctx, job.llm, recipient, insights, emotions, job.variation, templates)
	if err != nil {
		return nil, fmt.Errorf("template selection: %w", err)
	}

	tplDef := g.findTemplate(sel.TemplateID, templates)
	if tplDef == nil {
		return nil, fmt.Errorf("template not found: %s", sel.TemplateID)
	}

	copy, err := WriteCopy(ctx, job.llm, recipient, insights, emotions, job.variation, sel, *tplDef)
	if err != nil {
		return nil, fmt.Errorf("copywriting: %w", err)
	}

	dir := MergeArtDirection(sel, copy)

	palette, ok := GetNamedPalette(dir.PaletteName)
	if !ok {
		group := DetectEmotionGroup(emotions)
		palette = GetEmotionPalette(group)
	}

	occasionKey := string(DetectOccasion(recipient.Occasion))
	content := domain.CardContent{
		Headline:    dir.Headline,
		Body:        dir.Body,
		Closing:     dir.Closing,
		Signature:   dir.Signature,
		Recipient:   recipient.Name,
		Emotions:    emotions,
		OccasionKey: occasionKey,
	}

	illustrations := make(map[string]string)
	var illustration *domain.CardIllustration

	if dir.GenerateIllustration && dir.IllustrationPrompt != "" && g.assetLib != nil {
		slot := dir.IllustrationSlot
		if slot == "" {
			slot = "hero"
		}
		emotionNames := make([]string, len(emotions))
		for i, e := range emotions {
			emotionNames[i] = e.Name
		}
		imgBase64, imgErr := g.assetLib.GetOrGenerate(ctx, dir.IllustrationPrompt, occasionKey, emotionNames, slot, tplDef.Canvas.Width, tplDef.Canvas.Height)
		if imgErr != nil {
			log.Printf("illustration generation failed: %v", imgErr)
		} else {
			illustrations[slot] = imgBase64
			illustration = &domain.CardIllustration{
				PNGBase64: imgBase64,
				Slot:      slot,
				Prompt:    dir.IllustrationPrompt,
			}
		}
	}

	html, err := g.compiler.Compile(CompileInput{
		Template:      *tplDef,
		Palette:       palette,
		Content:       content,
		Illustrations: illustrations,
		DataFields:    make(map[string]string),
		Photos:        make(map[string]string),
		FontChoices:   dir.FontChoices,
		BackgroundIdx: dir.BackgroundChoice,
		Seed:          time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("compile template: %w", err)
	}

	result, err := g.renderer.RenderPNG(html, tplDef.Canvas.Width, tplDef.Canvas.Height)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	genMs := time.Since(start).Milliseconds()

	render := &domain.CardRender{
		PreviewPNG:   result,
		RecipeID:     dir.TemplateID,
		PaletteName:  dir.PaletteName,
		Content:      content,
		Model:        job.model,
		Illustration: illustration,
		Meta: &domain.CardMeta{
			Model:          job.model,
			GenerationMs:   genMs,
			TemplateTier:   tplDef.Tier,
			TemplateFamily: tplDef.Family,
			Orientation:    tplDef.Canvas.Orientation,
			Variation:      job.variation,
		},
	}

	validation := ValidateCard(render, *tplDef, palette)
	render.Meta.Validation = &validation

	return render, nil
}

func (g *CardGenerator) GenerateMemoryCard(ctx context.Context, llm port.LLMClient, recipient domain.RecipientDetails, insights []domain.PersonalityInsight, emotions []domain.EmotionSignal, chunks []domain.Chunk) *domain.CardRender {
	if g.renderer == nil || g.compiler == nil || g.tplStore == nil {
		return nil
	}

	memories, err := ExtractMemories(ctx, llm, chunks, recipient)
	if err != nil {
		log.Printf("memory extraction failed: %v", err)
		return nil
	}

	tplDef, err := g.tplStore.Get(ctx, "memory-evidence")
	if err != nil {
		log.Printf("memory-evidence template not found: %v", err)
		return nil
	}

	sel := TemplateSelection{
		TemplateID:           "memory-evidence",
		PaletteName:          pickPaletteForEmotions(emotions),
		GenerateIllustration: false,
	}

	copy, err := WriteCopy(ctx, llm, recipient, insights, emotions, "warm and nostalgic", sel, *tplDef)
	if err != nil {
		log.Printf("memory card copywriting failed: %v", err)
		return nil
	}

	palette, ok := GetNamedPalette(sel.PaletteName)
	if !ok {
		palette = GetEmotionPalette(DetectEmotionGroup(emotions))
	}

	content := domain.CardContent{
		Headline:    copy.Headline,
		Body:        buildMemoryBody(memories),
		Closing:     copy.Closing,
		Signature:   copy.Signature,
		Recipient:   recipient.Name,
		Emotions:    emotions,
		OccasionKey: string(DetectOccasion(recipient.Occasion)),
	}

	html, err := g.compiler.Compile(CompileInput{
		Template:      *tplDef,
		Palette:       palette,
		Content:       content,
		Illustrations: make(map[string]string),
		DataFields:    make(map[string]string),
		Photos:        make(map[string]string),
		FontChoices:   make(map[string]string),
		Seed:          time.Now().UnixNano(),
	})
	if err != nil {
		log.Printf("memory card compile failed: %v", err)
		return nil
	}

	result, err := g.renderer.RenderPNG(html, tplDef.Canvas.Width, tplDef.Canvas.Height)
	if err != nil {
		log.Printf("memory card render failed: %v", err)
		return nil
	}

	return &domain.CardRender{
		PreviewPNG:  result,
		RecipeID:    "memory-evidence",
		PaletteName: sel.PaletteName,
		Content:     content,
		Model:       "openai",
		CardType:    "memory_evidence",
		Evidences:   memories,
		Meta: &domain.CardMeta{
			Model:          "openai",
			TemplateTier:   tplDef.Tier,
			TemplateFamily: tplDef.Family,
			Orientation:    tplDef.Canvas.Orientation,
			Variation:      "memory evidence",
		},
	}
}

func pickPaletteForEmotions(emotions []domain.EmotionSignal) string {
	group := DetectEmotionGroup(emotions)
	for _, p := range namedPalettes {
		for _, e := range p.Emotions {
			if e == group {
				return p.Name
			}
		}
	}
	return "sunrise_warmth"
}

func buildMemoryBody(memories []domain.MemoryEvidence) string {
	if len(memories) == 0 {
		return ""
	}
	parts := make([]string, 0, len(memories))
	for _, m := range memories {
		parts = append(parts, fmt.Sprintf("\"%s\"", m.Quote))
	}
	body := strings.Join(parts, " ... ")
	if len(body) > 240 {
		body = body[:237] + "..."
	}
	return body
}

func (g *CardGenerator) findTemplate(id string, templates []domain.TemplateDefinition) *domain.TemplateDefinition {
	for i := range templates {
		if templates[i].ID == id {
			return &templates[i]
		}
	}
	return nil
}

func (g *CardGenerator) RenderPDF(ctx context.Context, card *domain.CardRender) (string, error) {
	return g.renderPDFMultiPage(ctx, card, false)
}

func (g *CardGenerator) RenderMultiPagePDF(ctx context.Context, card *domain.CardRender) (string, error) {
	return g.renderPDFMultiPage(ctx, card, true)
}

func (g *CardGenerator) renderPDFMultiPage(ctx context.Context, card *domain.CardRender, multiPage bool) (string, error) {
	if g.renderer == nil || g.compiler == nil {
		return "", fmt.Errorf("rendering not available")
	}

	palette, ok := GetNamedPalette(card.PaletteName)
	if !ok {
		group := DetectEmotionGroup(card.Content.Emotions)
		palette = GetEmotionPalette(group)
	}

	illustrations := make(map[string]string)
	if card.Illustration != nil && card.Illustration.PNGBase64 != "" {
		illustrations[card.Illustration.Slot] = card.Illustration.PNGBase64
	}

	tplDef, err := g.tplStore.Get(ctx, card.RecipeID)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}

	frontHTML, err := g.compiler.Compile(CompileInput{
		Template:      *tplDef,
		Palette:       palette,
		Content:       card.Content,
		Illustrations: illustrations,
		DataFields:    card.DataFields,
		Photos:        card.Photos,
		FontChoices:   make(map[string]string),
		Seed:          time.Now().UnixNano(),
	})
	if err != nil {
		return "", fmt.Errorf("compile front: %w", err)
	}

	if !multiPage {
		return g.renderer.RenderPrintPDF(frontHTML, tplDef.Canvas.Width, tplDef.Canvas.Height)
	}

	insideHTML, err := g.compiler.CompileInsidePage(CompileInput{
		Template: *tplDef,
		Palette:  palette,
		Content:  card.Content,
	})
	if err != nil {
		return "", fmt.Errorf("compile inside: %w", err)
	}

	return g.renderer.RenderMultiPagePDF(frontHTML, insideHTML, tplDef.Canvas.Width, tplDef.Canvas.Height)
}

func (g *CardGenerator) RenderPreview(ctx context.Context, card *domain.CardRender) (string, error) {
	if g.renderer == nil || g.compiler == nil {
		return "", fmt.Errorf("rendering not available")
	}

	palette, ok := GetNamedPalette(card.PaletteName)
	if !ok {
		group := DetectEmotionGroup(card.Content.Emotions)
		palette = GetEmotionPalette(group)
	}

	illustrations := make(map[string]string)
	if card.Illustration != nil && card.Illustration.PNGBase64 != "" {
		illustrations[card.Illustration.Slot] = card.Illustration.PNGBase64
	}

	tplDef, err := g.tplStore.Get(ctx, card.RecipeID)
	if err != nil {
		return "", fmt.Errorf("template not found: %w", err)
	}

	html, err := g.compiler.Compile(CompileInput{
		Template:      *tplDef,
		Palette:       palette,
		Content:       card.Content,
		Illustrations: illustrations,
		DataFields:    card.DataFields,
		Photos:        card.Photos,
		FontChoices:   make(map[string]string),
		Seed:          time.Now().UnixNano(),
	})
	if err != nil {
		return "", fmt.Errorf("compile: %w", err)
	}

	return g.renderer.RenderPNG(html, tplDef.Canvas.Width, tplDef.Canvas.Height)
}
