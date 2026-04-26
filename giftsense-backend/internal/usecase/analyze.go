package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/giftsense/backend/internal/adapter/cardrender"
	"github.com/giftsense/backend/internal/domain"
	"github.com/giftsense/backend/internal/port"
)

type AnalyzerConfig struct {
	MaxProcessedMessages int
	ChunkWindowSize      int
	ChunkOverlapSize     int
	TopK                 int
	NumRetrievalQueries  int
}

type LinkGeneratorFunc func(name string, budget domain.BudgetRange) domain.ShoppingLinks

type Analyzer struct {
	embedder     port.Embedder
	llm          port.LLMClient
	anthropicLLM port.LLMClient
	store        port.VectorStore
	linkGen      LinkGeneratorFunc
	cfg          AnalyzerConfig
	cardGen      *CardGenerator
}

func NewAnalyzer(embedder port.Embedder, llm port.LLMClient, store port.VectorStore, linkGen LinkGeneratorFunc, cfg AnalyzerConfig, renderer *cardrender.Renderer, engine *cardrender.TemplateEngine, anthropicLLM port.LLMClient, imageGen port.ImageGenerator, assetLib *AssetLibrary) *Analyzer {
	return &Analyzer{embedder: embedder, llm: llm, anthropicLLM: anthropicLLM, store: store, linkGen: linkGen, cfg: cfg, cardGen: NewCardGenerator(llm, anthropicLLM, renderer, engine, imageGen, assetLib)}
}

func (a *Analyzer) SetTemplateStore(store port.TemplateStore, compiler *HTMLCompiler) {
	a.cardGen.SetTemplateStore(store, compiler)
}

type llmResponse struct {
	PersonalityInsights []domain.PersonalityInsight `json:"personality_insights"`
	GiftSuggestions     []llmGiftSuggestion         `json:"gift_suggestions"`
}

type llmGiftSuggestion struct {
	Name              string `json:"name"`
	Reason            string `json:"reason"`
	EstimatedPriceINR string `json:"estimated_price_inr"`
	Category          string `json:"category"`
}

func (a *Analyzer) Analyze(ctx context.Context, sessionID, conversationText string, recipient domain.RecipientDetails) (domain.AnalysisResult, error) {
	messages, err := ParseConversation(conversationText, a.cfg.MaxProcessedMessages)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("parse: %w", err)
	}
	anonMessages, _, err := AnonymizeMessages(messages)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("anonymize: %w", err)
	}
	chunks, err := ChunkMessages(sessionID, anonMessages, a.cfg.ChunkWindowSize, a.cfg.ChunkOverlapSize)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("chunk: %w", err)
	}
	chunksByID, err := a.embedAndStore(ctx, sessionID, chunks)
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	defer func() { _ = a.store.DeleteSession(context.Background(), sessionID) }()

	queries := BuildRetrievalQueries(recipient, a.cfg.NumRetrievalQueries)
	retrieved, err := RetrieveAndRerank(ctx, sessionID, queries, chunksByID, a.embedder, a.store, a.cfg.TopK)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("retrieve: %w", err)
	}
	if len(retrieved) == 0 {
		return domain.AnalysisResult{}, domain.ErrRetrievalFailed
	}

	var (
		result   domain.AnalysisResult
		emotions []domain.EmotionSignal
	)
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		ctx10, cancel := context.WithTimeout(egCtx, 10*time.Second)
		defer cancel()
		var extractErr error
		emotions, extractErr = a.ExtractChatEmotions(ctx10, retrieved, recipient)
		if extractErr != nil {
			emotions = []domain.EmotionSignal{{Name: "Warmth", Emoji: "🤗", Intensity: 0.7}}
		}
		return nil
	})

	eg.Go(func() error {
		prompt := BuildAnalysisPrompt(retrieved, recipient)
		raw, llmErr := a.llm.Complete(egCtx, prompt, port.CompletionOptions{JSONMode: true})
		if llmErr != nil {
			return fmt.Errorf("llm complete: %w", llmErr)
		}
		var parseErr error
		result, parseErr = parseAndEnrichLLMResponse(raw, recipient.Budget, a.linkGen)
		return parseErr
	})

	if err := eg.Wait(); err != nil {
		return domain.AnalysisResult{}, err
	}

	dataFields := ExtractDataFields(messages, recipient)
	result.DataFields = dataFields
	result.Cards = a.cardGen.Generate(ctx, recipient, result.PersonalityInsights, emotions)
	for _, card := range result.Cards {
		card.DataFields = dataFields
	}

	if len(retrieved) > 0 && a.llm != nil {
		memCard := a.cardGen.GenerateMemoryCard(ctx, a.llm, recipient, result.PersonalityInsights, emotions, retrieved)
		if memCard != nil {
			memCard.DataFields = dataFields
			result.Cards = append(result.Cards, memCard)
		}
	}

	return result, nil
}

func ExtractDataFields(messages []domain.Message, recipient domain.RecipientDetails) map[string]string {
	fields := map[string]string{
		"recipient_name": recipient.Name,
		"occasion":       recipient.Occasion,
	}
	if len(messages) > 0 {
		fields["message_count"] = fmt.Sprintf("%d", len(messages))
	}

	emojiCounts := make(map[string]int)
	for _, m := range messages {
		for _, r := range m.Text {
			if isEmoji(r) {
				emojiCounts[string(r)]++
			}
		}
	}
	if len(emojiCounts) > 0 {
		topEmoji := ""
		topCount := 0
		for e, c := range emojiCounts {
			if c > topCount {
				topEmoji = e
				topCount = c
			}
		}
		fields["top_emoji"] = topEmoji
	}

	return fields
}

func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) ||
		(r >= 0x1F300 && r <= 0x1F5FF) ||
		(r >= 0x1F680 && r <= 0x1F6FF) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x2600 && r <= 0x26FF) ||
		(r >= 0x2700 && r <= 0x27BF) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		(r >= 0x1FA00 && r <= 0x1FA6F) ||
		(r >= 0x1FA70 && r <= 0x1FAFF) ||
		(r >= 0x2702 && r <= 0x27B0) ||
		r == 0x200D || r == 0xFE0F ||
		(r >= 0x1F1E0 && r <= 0x1F1FF) ||
		r == 0x2764 || r == 0x2763
}

func (a *Analyzer) embedAndStore(ctx context.Context, sessionID string, chunks []domain.Chunk) (map[string]domain.Chunk, error) {
	texts := make([]string, len(chunks))
	for i, c := range chunks {
		texts[i] = c.AnonymizedText
	}
	vectors, err := a.embedder.Embed(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed chunks: %w", err)
	}
	if err := a.store.Upsert(ctx, sessionID, chunks, vectors); err != nil {
		return nil, fmt.Errorf("upsert vectors: %w", err)
	}
	byID := make(map[string]domain.Chunk, len(chunks))
	for _, c := range chunks {
		byID[c.ID] = c
	}
	return byID, nil
}

func BuildAnalysisPrompt(chunks []domain.Chunk, recipient domain.RecipientDetails) string {
	var sb strings.Builder
	budgetDesc := fmt.Sprintf("₹%d+", recipient.Budget.MinINR)
	if recipient.Budget.MaxINR > 0 {
		budgetDesc = fmt.Sprintf("₹%d–₹%d", recipient.Budget.MinINR, recipient.Budget.MaxINR)
	}
	fmt.Fprintf(&sb, "Find gift ideas for %s", recipient.Name)
	if recipient.Relation != "" {
		fmt.Fprintf(&sb, " (%s)", recipient.Relation)
	}
	fmt.Fprintf(&sb, ", occasion: %s, budget: %s (%s).\n\n", recipient.Occasion, recipient.Budget.Tier, budgetDesc)
	sb.WriteString("Conversation excerpts:\n\n")
	for _, c := range chunks {
		sb.WriteString("---\n")
		sb.WriteString(c.AnonymizedText)
		sb.WriteString("\n")
	}
	sb.WriteString("\nSuggest 3-5 specific, thoughtful gifts. Each estimated_price_inr must be within the stated budget.")
	return sb.String()
}

// AnalyzeFromTranscript runs the gift analysis pipeline starting from a prose transcript,
// bypassing the WhatsApp parser. Used by the audio flow after transcription.
func (a *Analyzer) AnalyzeFromTranscript(ctx context.Context, sessionID, transcript string, recipient domain.RecipientDetails, confirmedEmotions []domain.EmotionSignal) (domain.AnalysisResult, error) {
	chunks := ChunkTranscript(sessionID, transcript, 200, 50)
	if len(chunks) == 0 {
		return domain.AnalysisResult{}, domain.ErrConversationTooShort
	}

	chunksByID, err := a.embedAndStore(ctx, sessionID, chunks)
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	defer func() { _ = a.store.DeleteSession(context.Background(), sessionID) }()

	queries := BuildRetrievalQueries(recipient, a.cfg.NumRetrievalQueries)
	retrieved, err := RetrieveAndRerank(ctx, sessionID, queries, chunksByID, a.embedder, a.store, a.cfg.TopK)
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("retrieve: %w", err)
	}
	if len(retrieved) == 0 {
		return domain.AnalysisResult{}, domain.ErrRetrievalFailed
	}

	prompt := BuildAnalysisPromptWithEmotions(retrieved, recipient, confirmedEmotions)
	raw, err := a.llm.Complete(ctx, prompt, port.CompletionOptions{JSONMode: true})
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("llm complete: %w", err)
	}
	result, err := parseAndEnrichLLMResponse(raw, recipient.Budget, a.linkGen)
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	dataFields := map[string]string{
		"recipient_name": recipient.Name,
		"occasion":       recipient.Occasion,
	}
	result.DataFields = dataFields
	result.Cards = a.cardGen.Generate(ctx, recipient, result.PersonalityInsights, confirmedEmotions)
	for _, card := range result.Cards {
		card.DataFields = dataFields
	}

	if len(retrieved) > 0 && a.llm != nil {
		memCard := a.cardGen.GenerateMemoryCard(ctx, a.llm, recipient, result.PersonalityInsights, confirmedEmotions, retrieved)
		if memCard != nil {
			memCard.DataFields = dataFields
			result.Cards = append(result.Cards, memCard)
		}
	}

	return result, nil
}

// BuildAnalysisPromptWithEmotions builds the gift analysis prompt, optionally appending
// song emotion context when the input came from a song.
func BuildAnalysisPromptWithEmotions(chunks []domain.Chunk, recipient domain.RecipientDetails, emotions []domain.EmotionSignal) string {
	base := BuildAnalysisPrompt(chunks, recipient)
	if len(emotions) == 0 {
		return base
	}

	var parts []string
	for _, e := range emotions {
		parts = append(parts, fmt.Sprintf("%s %s (%.1f)", e.Name, e.Emoji, e.Intensity))
	}

	return base + fmt.Sprintf(
		"\n\nNote: These gift suggestions were inspired by a song the user chose for this person.\nThe song's emotional fingerprint: [%s].\nLet this emotional warmth inform your suggestions.",
		strings.Join(parts, ", "),
	)
}

// AnalyzeFromSongEmotions generates gift suggestions based on a Spotify song's emotions
// and recipient details. This skips the RAG pipeline entirely since there is no
// conversation text — the LLM prompt is built directly from the song context.
func (a *Analyzer) AnalyzeFromSongEmotions(ctx context.Context, recipient domain.RecipientDetails, songName, artist string, emotions []domain.EmotionSignal) (domain.AnalysisResult, error) {
	prompt := buildSongGiftPrompt(recipient, songName, artist, emotions)
	raw, err := a.llm.Complete(ctx, prompt, port.CompletionOptions{JSONMode: true})
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("llm complete (song): %w", err)
	}
	result, err := parseAndEnrichLLMResponse(raw, recipient.Budget, a.linkGen)
	if err != nil {
		return domain.AnalysisResult{}, err
	}
	dataFields := map[string]string{
		"recipient_name": recipient.Name,
		"occasion":       recipient.Occasion,
		"song_name":      songName,
		"artist_name":    artist,
	}
	result.DataFields = dataFields
	result.Cards = a.cardGen.Generate(ctx, recipient, result.PersonalityInsights, emotions)
	for _, card := range result.Cards {
		card.DataFields = dataFields
	}
	return result, nil
}

func buildSongGiftPrompt(recipient domain.RecipientDetails, songName, artist string, emotions []domain.EmotionSignal) string {
	var sb strings.Builder

	budgetDesc := fmt.Sprintf("₹%d+", recipient.Budget.MinINR)
	if recipient.Budget.MaxINR > 0 {
		budgetDesc = fmt.Sprintf("₹%d–₹%d", recipient.Budget.MinINR, recipient.Budget.MaxINR)
	}

	fmt.Fprintf(&sb, "Find gift ideas for %s", recipient.Name)
	if recipient.Relation != "" {
		fmt.Fprintf(&sb, " (%s)", recipient.Relation)
	}
	fmt.Fprintf(&sb, ", occasion: %s, budget: %s (%s).\n\n", recipient.Occasion, recipient.Budget.Tier, budgetDesc)

	fmt.Fprintf(&sb, "The user has dedicated the song \"%s\" by %s to express their feelings for this person.\n", songName, artist)

	if len(emotions) > 0 {
		var parts []string
		for _, e := range emotions {
			parts = append(parts, fmt.Sprintf("%s %s (%.1f)", e.Name, e.Emoji, e.Intensity))
		}
		fmt.Fprintf(&sb, "The song's emotional fingerprint: [%s].\n", strings.Join(parts, ", "))
	}

	sb.WriteString("\nSuggest 3-5 specific, thoughtful gifts that resonate with the emotions this song conveys. Each estimated_price_inr must be within the stated budget.")
	return sb.String()
}

func parseAndEnrichLLMResponse(raw string, budget domain.BudgetRange, linkGen LinkGeneratorFunc) (domain.AnalysisResult, error) {
	var resp llmResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("%w: %s", domain.ErrLLMResponseInvalid, err.Error())
	}
	if len(resp.GiftSuggestions) == 0 {
		return domain.AnalysisResult{}, domain.ErrLLMResponseInvalid
	}
	suggestions := make([]domain.GiftSuggestion, 0, len(resp.GiftSuggestions))
	for _, s := range resp.GiftSuggestions {
		if s.Name == "" {
			continue
		}
		suggestions = append(suggestions, domain.GiftSuggestion{
			Name:              s.Name,
			Reason:            s.Reason,
			EstimatedPriceINR: s.EstimatedPriceINR,
			Category:          s.Category,
			Links:             linkGen(s.Name, budget),
		})
	}
	if len(suggestions) == 0 {
		return domain.AnalysisResult{}, domain.ErrAllSuggestionsFiltered
	}
	return domain.AnalysisResult{PersonalityInsights: resp.PersonalityInsights, GiftSuggestions: suggestions}, nil
}
