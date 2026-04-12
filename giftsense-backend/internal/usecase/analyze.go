package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	embedder port.Embedder
	llm      port.LLMClient
	store    port.VectorStore
	linkGen  LinkGeneratorFunc
	cfg      AnalyzerConfig
}

func NewAnalyzer(embedder port.Embedder, llm port.LLMClient, store port.VectorStore, linkGen LinkGeneratorFunc, cfg AnalyzerConfig) *Analyzer {
	return &Analyzer{embedder: embedder, llm: llm, store: store, linkGen: linkGen, cfg: cfg}
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
	prompt := BuildAnalysisPrompt(retrieved, recipient)
	raw, err := a.llm.Complete(ctx, prompt, port.CompletionOptions{JSONMode: true})
	if err != nil {
		return domain.AnalysisResult{}, fmt.Errorf("llm complete: %w", err)
	}
	return parseAndEnrichLLMResponse(raw, recipient.Budget, a.linkGen)
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
	return parseAndEnrichLLMResponse(raw, recipient.Budget, a.linkGen)
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
	return parseAndEnrichLLMResponse(raw, recipient.Budget, a.linkGen)
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
