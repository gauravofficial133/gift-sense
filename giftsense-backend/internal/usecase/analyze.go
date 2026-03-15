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
	sb.WriteString(fmt.Sprintf("Find gift ideas for %s", recipient.Name))
	if recipient.Relation != "" {
		sb.WriteString(fmt.Sprintf(" (%s)", recipient.Relation))
	}
	sb.WriteString(fmt.Sprintf(", occasion: %s, budget: %s (%s).\n\n", recipient.Occasion, recipient.Budget.Tier, budgetDesc))
	sb.WriteString("Conversation excerpts:\n\n")
	for _, c := range chunks {
		sb.WriteString("---\n")
		sb.WriteString(c.AnonymizedText)
		sb.WriteString("\n")
	}
	sb.WriteString("\nSuggest 3-5 specific, thoughtful gifts. Each estimated_price_inr must be within the stated budget.")
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
